/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

// Package view implements the default `ailog` route: read a log source,
// filter it, and print or page the matching records.
package view

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/os-artificer/AiT/internal/log/queryflags"
	"github.com/os-artificer/AiT/internal/pkg/logsrc"
	"github.com/os-artificer/AiT/internal/pkg/pager"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	// noMatchMessage is printed to stderr when a query matched nothing, so that
	// stdout stays empty for pipelines.
	noMatchMessage = toolname.AILog + ": no matching log entries"

	// produceBuffer lets the reader run ahead of the pager while the user is
	// looking at a frame.
	produceBuffer = 256
)

// Run executes the default view/search route with argv (flags and keywords).
// Returns a flag parsing error, a query validation error with Try hints, or a
// source error from logsrc.
func Run(ctx context.Context, stdout, stderr io.Writer, args []string) error {
	cmd := newCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	return cmd.Execute()
}

// RunSearch runs a query with already-parsed flags, used by `ailog search`.
// Returns the same errors as Run.
func RunSearch(ctx context.Context, stdout, stderr io.Writer, f queryflags.Flags, keywords []string) error {
	return runView(ctx, stdout, stderr, f, keywords)
}

func runView(ctx context.Context, stdout, stderr io.Writer, f queryflags.Flags, keywords []string) error {
	q, err := queryflags.BuildQuery(f, keywords, time.Now())
	if err != nil {
		return err
	}
	filt, err := logsrc.NewFilter(q)
	if err != nil {
		return err
	}

	resolver := logsrc.DefaultResolver()
	resolver.Notice = stderr
	if f.Quiet {
		resolver.Notice = io.Discard
	}

	src, cur, err := resolver.OpenResolved(ctx, q)
	if err != nil {
		return err
	}
	defer cur.Close()

	if q.Follow {
		return streamFollow(ctx, stdout, cur, filt, q, f)
	}
	if q.Head && queryflags.UsePager(f) {
		return pageProgressive(ctx, stdout, cur, filt, q, f)
	}
	return renderCollected(ctx, stdout, stderr, src, cur, filt, q, f)
}

// renderCollected finishes the scan first, which last-N queries require, then
// prints or pages the result.
func renderCollected(ctx context.Context, stdout, stderr io.Writer, src logsrc.Source, cur logsrc.Cursor,
	filt *logsrc.Filter, q logsrc.Query, f queryflags.Flags) error {

	collected, err := logsrc.Collect(ctx, src, cur, q, filt)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	if collected.Window > 0 {
		logsrc.WriteNotice(stderr, f.Quiet, "searched the last %s; use --since to widen the range",
			logsrc.FormatWindow(collected.Window))
	}
	if len(collected.Records) == 0 {
		if !f.Quiet {
			_, _ = fmt.Fprintln(stderr, noMatchMessage)
		}
		return nil
	}

	lines := make([]string, 0, len(collected.Records))
	for _, rec := range collected.Records {
		lines = append(lines, formatRecord(rec, q, f.JSON))
	}
	if !queryflags.UsePager(f) {
		return writeLines(stdout, lines)
	}
	return pager.Page(ctx, stdout, &pager.SliceSource{Lines: lines}, pager.Options{Bounded: true})
}

// pageProgressive pages a --head query while it is still being read, so the
// first screen appears without waiting for the whole scan.
//
// The producer must stop before this returns: the caller closes the cursor as
// soon as it does, and reading a closed cursor is not safe.
func pageProgressive(ctx context.Context, stdout io.Writer, cur logsrc.Cursor,
	filt *logsrc.Filter, q logsrc.Query, f queryflags.Flags) error {

	produce, stopProducing := context.WithCancel(ctx)
	defer stopProducing()

	lines := make(chan string, produceBuffer)
	src := &pager.ChanSource{Lines: lines}
	done := make(chan struct{})
	go func() {
		defer close(done)
		produceLines(produce, lines, src, cur, filt, q, f)
	}()

	err := pager.Page(ctx, stdout, src, pager.Options{StatusHint: pager.HintLoading})

	stopProducing()
	// Close the cursor so a blocked cur.Next (native tool pipe) unblocks; the
	// caller also defers Close, and Close is idempotent for exec/file cursors.
	_ = cur.Close()
	<-done
	return err
}

// produceLines feeds formatted matches to the pager until the limit is reached,
// the cursor ends, or ctx is cancelled.
func produceLines(ctx context.Context, out chan<- string, src *pager.ChanSource, cur logsrc.Cursor,
	filt *logsrc.Filter, q logsrc.Query, f queryflags.Flags) {

	defer close(out)

	limit := q.EffectiveLimit()
	sent := 0
	var rec logsrc.Record
	for sent < limit && cur.Next(&rec) {
		if filt != nil && !filt.Match(rec) {
			continue
		}
		select {
		case out <- formatRecord(rec, q, f.JSON):
			sent++
		case <-ctx.Done():
			return
		}
	}
	if err := cur.Err(); err != nil && ctx.Err() == nil {
		src.Fail(err)
	}
}

// streamFollow prints matches as they arrive; --follow never pages.
func streamFollow(ctx context.Context, stdout io.Writer, cur logsrc.Cursor,
	filt *logsrc.Filter, q logsrc.Query, f queryflags.Flags) error {

	var rec logsrc.Record
	for {
		if ctx.Err() != nil {
			return nil
		}
		if !cur.Next(&rec) {
			if ctx.Err() != nil {
				return nil
			}
			return cur.Err()
		}
		if filt != nil && !filt.Match(rec) {
			continue
		}
		if _, err := fmt.Fprintln(stdout, formatRecord(rec, q, f.JSON)); err != nil {
			return err
		}
	}
}

func writeLines(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}
