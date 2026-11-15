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

package shared

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// WatchSeparator separates human --watch samples (JSON watch omits it).
const WatchSeparator = "---\n"

// BindJSONInterval adds --json, --interval, and --watch to cmd.
func BindJSONInterval(cmd *cobra.Command, f *Flags) {
	cmd.Flags().BoolVar(&f.JSON, "json", false, "emit JSON")
	cmd.Flags().DurationVar(&f.Interval, "interval", time.Second, "sample interval for --watch")
	cmd.Flags().BoolVar(&f.Watch, "watch", false, "refresh until interrupted")
}

// BindProcessQuery adds --limit (default DefaultProcessLimit) and --sort.
func BindProcessQuery(cmd *cobra.Command, f *Flags) {
	BindProcessQueryLimit(cmd, f, hoststats.DefaultProcessLimit)
}

// BindProcessQueryLimit adds --limit with defaultLimit and --sort.
// defaultLimit 0 leaves Limit unset so callers can treat 0 as all processes (aihost top).
func BindProcessQueryLimit(cmd *cobra.Command, f *Flags, defaultLimit int) {
	cmd.Flags().IntVar(&f.Limit, "limit", defaultLimit, "max process rows")
	cmd.Flags().StringVar(&f.Sort, "sort", hoststats.SortCPU, "primary sort: cpu|mem|net|disk")
}

// RunSection samples once (or watches) and writes one section identified by sectionID.
func RunSection(ctx context.Context, w io.Writer, sectionID string, f Flags) error {
	sec := hoststats.SectionByID(sectionID)
	if sec == nil {
		return fmt.Errorf("unknown section %q", sectionID)
	}
	pq, err := f.ProcessQuery()
	if err != nil {
		return err
	}
	col := hoststats.NewCollector()
	if !f.Watch {
		snap, err := col.Sample(ctx, pq)
		if err != nil {
			return err
		}
		if f.JSON {
			return hoststats.WriteJSON(w, snap, sec)
		}
		return hoststats.WriteHuman(w, snap, []hoststats.Section{sec})
	}
	return watch(ctx, w, col, pq, f, []hoststats.Section{sec})
}

// RunSnapshot samples and writes all default sections (or JSON full snap).
func RunSnapshot(ctx context.Context, w io.Writer, f Flags) error {
	pq, err := f.ProcessQuery()
	if err != nil {
		return err
	}
	col := hoststats.NewCollector()
	if !f.Watch {
		snap, err := col.Sample(ctx, pq)
		if err != nil {
			return err
		}
		if f.JSON {
			return hoststats.WriteJSON(w, snap, nil)
		}
		return hoststats.WriteHuman(w, snap, nil)
	}
	return watch(ctx, w, col, pq, f, nil)
}

// WriterIsTerminal reports whether w is an *os.File connected to a TTY.
func WriterIsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// WatchEvery calls writeOnce immediately, then again on each interval until ctx is done.
// writeOnce should write one sample; for human multi-line output callers append WatchSeparator themselves.
// Returns writeOnce errors; ctx cancel returns nil.
func WatchEvery(ctx context.Context, interval time.Duration, writeOnce func() error) error {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := writeOnce(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func watch(ctx context.Context, w io.Writer, col *hoststats.Collector, pq hoststats.ProcessQuery, f Flags, sections []hoststats.Section) error {
	return WatchEvery(ctx, f.Interval, func() error {
		snap, err := col.Sample(ctx, pq)
		if err != nil {
			return err
		}
		return writeWatchSample(w, snap, f, sections)
	})
}

func writeWatchSample(w io.Writer, snap hoststats.Snapshot, f Flags, sections []hoststats.Section) error {
	if f.JSON {
		return hoststats.WriteJSON(w, snap, sectionOrNil(sections))
	}
	if err := hoststats.WriteHuman(w, snap, sections); err != nil {
		return err
	}
	_, err := io.WriteString(w, WatchSeparator)
	return err
}

func sectionOrNil(sections []hoststats.Section) hoststats.Section {
	if len(sections) == 1 {
		return sections[0]
	}
	return nil
}
