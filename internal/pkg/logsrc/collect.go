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

package logsrc

import (
	"context"
	"time"
)

// Collected is the outcome of a bounded collection pass.
type Collected struct {
	// Records are in chronological order (oldest first).
	Records []Record
	// Window is the time span actually searched when the backend is
	// forward-only and had to bound the query; zero when the whole store
	// (or an explicit --since range) was searched.
	Window time.Duration
}

// Collect gathers the records for q from first, which must be the cursor
// OpenResolved returned for src.
//
// Head queries take the first N matches; reverse-capable sources stop after N
// matches; forward-only WindowedSource backends retry over growing time
// windows; everything else scans to EOF keeping the last N matches.
//
// The caller retains ownership of first and must Close it. Cursors that
// Collect opens for wider windows are closed before returning. Returns the
// cursor's error, or ctx errors surfaced by the cursor.
func Collect(ctx context.Context, src Source, first Cursor, q Query, filt *Filter) (Collected, error) {
	n := q.EffectiveLimit()

	if q.Head {
		recs, err := CollectHeadN(first, filt, n)
		return Collected{Records: recs}, err
	}
	if src != nil && src.SupportsReverse() {
		recs, err := CollectLastN(first, filt, n, true)
		return Collected{Records: recs}, err
	}

	ws, ok := src.(WindowedSource)
	if !ok || !q.Since.IsZero() {
		recs, err := CollectLastN(first, filt, n, false)
		return Collected{Records: recs}, err
	}
	windows := windowsFor(ws, q)
	if len(windows) == 0 {
		recs, err := CollectLastN(first, filt, n, false)
		return Collected{Records: recs}, err
	}
	return collectWindowed(ctx, ws, first, q, filt, n, windows)
}

// windowsFor returns the expanding window list for ws, preferring a
// query-aware list when the source implements WindowedSourceFor.
func windowsFor(ws WindowedSource, q Query) []time.Duration {
	if wf, ok := ws.(WindowedSourceFor); ok {
		return wf.WindowsFor(q)
	}
	return ws.Windows()
}

// collectWindowed walks windows from the narrowest to the widest until it
// has n matches. first covers windows[0]; later cursors are opened and
// closed here.
func collectWindowed(ctx context.Context, ws WindowedSource, first Cursor, q Query, filt *Filter, n int, windows []time.Duration) (Collected, error) {
	cur := first

	for i, window := range windows {
		recs, err := CollectLastN(cur, filt, n, false)
		if i > 0 {
			_ = cur.Close()
		}
		if err != nil {
			return Collected{Window: window}, err
		}
		if len(recs) >= n || i == len(windows)-1 {
			return Collected{Records: recs, Window: window}, nil
		}

		next, openErr := ws.OpenWindow(ctx, q, windows[i+1])
		if openErr != nil {
			// Keep what the narrower window produced rather than failing.
			return Collected{Records: recs, Window: window}, nil
		}
		cur = next
	}
	return Collected{}, nil
}

// CollectLastN reads cur until EOF, retaining the last n records that match
// filt. When reverse is true the cursor yields newest-first, so collection
// stops after n matches and the result is flipped back to chronological order.
// A nil filt matches every record. Returns cur.Err() if iteration failed.
func CollectLastN(cur Cursor, filt *Filter, n int, reverse bool) ([]Record, error) {
	if n <= 0 {
		return nil, cur.Err()
	}
	if reverse {
		return collectReverse(cur, filt, n)
	}

	ring := NewRing(n)
	var rec Record
	for cur.Next(&rec) {
		if filt != nil && !filt.Match(rec) {
			continue
		}
		ring.Add(rec)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return ring.Snapshot(), nil
}

// CollectHeadN reads cur until n records match filt or the cursor ends.
// A nil filt matches every record. Returns cur.Err() if iteration failed.
func CollectHeadN(cur Cursor, filt *Filter, n int) ([]Record, error) {
	if n <= 0 {
		return nil, cur.Err()
	}
	out := make([]Record, 0, n)
	var rec Record
	for cur.Next(&rec) {
		if filt != nil && !filt.Match(rec) {
			continue
		}
		out = append(out, rec)
		if len(out) >= n {
			return out, nil
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func collectReverse(cur Cursor, filt *Filter, n int) ([]Record, error) {
	out := make([]Record, 0, n)
	var rec Record
	for cur.Next(&rec) {
		if filt != nil && !filt.Match(rec) {
			continue
		}
		out = append(out, rec)
		if len(out) >= n {
			break
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	reverseRecords(out)
	return out, nil
}

func reverseRecords(recs []Record) {
	for i, j := 0, len(recs)-1; i < j; i, j = i+1, j-1 {
		recs[i], recs[j] = recs[j], recs[i]
	}
}
