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

package logsrc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

// windowedSource is a forward-only source whose store is only fully visible in
// the widest window, standing in for macOS `log show`.
type windowedSource struct {
	windows []time.Duration
	// perWindow maps a window to the records it can see.
	perWindow map[time.Duration][]logsrc.Record
	opened    []time.Duration
}

func (s *windowedSource) Name() string             { return "windowed" }
func (s *windowedSource) SupportsReverse() bool    { return false }
func (s *windowedSource) Windows() []time.Duration { return s.windows }

func (s *windowedSource) Open(ctx context.Context, q logsrc.Query) (logsrc.Cursor, error) {
	return s.OpenWindow(ctx, q, s.windows[0])
}

func (s *windowedSource) OpenWindow(ctx context.Context, q logsrc.Query, window time.Duration) (logsrc.Cursor, error) {
	s.opened = append(s.opened, window)
	return &sliceCur{recs: s.perWindow[window]}, nil
}

func records(msgs ...string) []logsrc.Record {
	out := make([]logsrc.Record, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, logsrc.Record{Message: m})
	}
	return out
}

// LG11: a forward-only windowed source widens its window until it has N
// matches, instead of silently reporting only what the narrowest window saw.
func TestCollect_WidensWindowUntilEnoughMatches(t *testing.T) {
	src := &windowedSource{
		windows: []time.Duration{time.Hour, 6 * time.Hour, 24 * time.Hour},
		perWindow: map[time.Duration][]logsrc.Record{
			time.Hour:      records("a"),
			6 * time.Hour:  records("a", "b"),
			24 * time.Hour: records("a", "b", "c"),
		},
	}
	first, err := src.Open(context.Background(), logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()

	got, err := logsrc.Collect(context.Background(), src, first, logsrc.Query{Limit: 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 3 {
		t.Fatalf("records=%d want 3 (window expansion did not happen)", len(got.Records))
	}
	if got.Window != 24*time.Hour {
		t.Fatalf("window=%v want 24h", got.Window)
	}
	if len(src.opened) != 3 {
		t.Fatalf("opened=%v want all three windows tried", src.opened)
	}
}

// LG12: expansion stops as soon as a window yields enough matches.
func TestCollect_StopsAtFirstSufficientWindow(t *testing.T) {
	src := &windowedSource{
		windows: []time.Duration{time.Hour, 6 * time.Hour},
		perWindow: map[time.Duration][]logsrc.Record{
			time.Hour:     records("a", "b"),
			6 * time.Hour: records("a", "b", "c"),
		},
	}
	first, err := src.Open(context.Background(), logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()

	got, err := logsrc.Collect(context.Background(), src, first, logsrc.Query{Limit: 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 2 || got.Window != time.Hour {
		t.Fatalf("records=%d window=%v want 2 records in 1h", len(got.Records), got.Window)
	}
	if len(src.opened) != 1 {
		t.Fatalf("opened=%v want only the narrowest window", src.opened)
	}
}

// LG13: an explicit --since disables windowing, because the user asked for a
// specific range and must not get a narrower one.
func TestCollect_ExplicitSinceDisablesWindowing(t *testing.T) {
	src := &windowedSource{
		windows: []time.Duration{time.Hour, 6 * time.Hour},
		perWindow: map[time.Duration][]logsrc.Record{
			time.Hour: records("a"),
		},
	}
	first, err := src.Open(context.Background(), logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()

	q := logsrc.Query{Limit: 5, Since: time.Now().Add(-72 * time.Hour)}
	got, err := logsrc.Collect(context.Background(), src, first, q, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Window != 0 {
		t.Fatalf("window=%v want 0 so no misleading notice is printed", got.Window)
	}
	if len(src.opened) != 1 {
		t.Fatalf("opened=%v want no expansion", src.opened)
	}
}

// LG14: a reverse-capable source stops after N matches rather than draining.
func TestCollect_ReverseSourceStopsAfterN(t *testing.T) {
	src := reverseSource{}
	cur := &sliceCur{recs: records("d", "c", "b", "a")}

	got, err := logsrc.Collect(context.Background(), src, cur, logsrc.Query{Limit: 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 2 || got.Records[0].Message != "c" || got.Records[1].Message != "d" {
		t.Fatalf("%+v want chronological [c d]", got.Records)
	}
	if cur.i != 2 {
		t.Fatalf("consumed %d records, want 2 (should stop early)", cur.i)
	}
}

// LG15: --head takes the first N matches in stream order.
func TestCollect_HeadTakesFirstMatches(t *testing.T) {
	cur := &sliceCur{recs: records("a", "b", "c")}
	filt, err := logsrc.NewFilter(logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}

	got, err := logsrc.Collect(context.Background(), forwardSource{}, cur, logsrc.Query{Head: true, Limit: 2}, filt)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 2 || got.Records[0].Message != "a" {
		t.Fatalf("%+v want [a b]", got.Records)
	}
}

// LG16: the last-N cut runs after filtering, so -n counts matches and not
// records read from the backend.
func TestCollect_LimitCountsMatchesNotRecords(t *testing.T) {
	cur := &sliceCur{recs: records("miss", "hit one", "miss", "hit two")}
	filt, err := logsrc.NewFilter(logsrc.Query{Keywords: []string{"hit"}})
	if err != nil {
		t.Fatal(err)
	}

	got, err := logsrc.Collect(context.Background(), forwardSource{}, cur, logsrc.Query{Limit: 2}, filt)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 2 {
		t.Fatalf("%+v want both matches", got.Records)
	}
}

// LG17: a cursor error surfaces instead of being reported as an empty result.
func TestCollect_PropagatesCursorError(t *testing.T) {
	want := errors.New("read failed")
	cur := &sliceCur{recs: records("a"), err: want}

	_, err := logsrc.Collect(context.Background(), forwardSource{}, cur, logsrc.Query{Limit: 5}, nil)
	if !errors.Is(err, want) {
		t.Fatalf("err=%v want %v", err, want)
	}
}

// LG18: a failure to open a wider window keeps the narrower window's matches
// rather than losing them.
func TestCollect_WindowOpenFailureKeepsPartialResult(t *testing.T) {
	src := &failingWindowSource{windowedSource{
		windows: []time.Duration{time.Hour, 6 * time.Hour},
		perWindow: map[time.Duration][]logsrc.Record{
			time.Hour: records("a"),
		},
	}}
	first, err := src.Open(context.Background(), logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()

	got, err := logsrc.Collect(context.Background(), src, first, logsrc.Query{Limit: 5}, nil)
	if err != nil {
		t.Fatalf("want partial success, got %v", err)
	}
	if len(got.Records) != 1 {
		t.Fatalf("%+v want the one record the narrow window found", got.Records)
	}
}

type failingWindowSource struct{ windowedSource }

func (s *failingWindowSource) OpenWindow(ctx context.Context, q logsrc.Query, window time.Duration) (logsrc.Cursor, error) {
	if window != s.windows[0] {
		return nil, fmt.Errorf("cannot widen")
	}
	return s.windowedSource.OpenWindow(ctx, q, window)
}

type forwardSource struct{}

func (forwardSource) Name() string          { return "forward" }
func (forwardSource) SupportsReverse() bool { return false }
func (forwardSource) Open(context.Context, logsrc.Query) (logsrc.Cursor, error) {
	return nil, errors.New("not used")
}

type reverseSource struct{}

func (reverseSource) Name() string          { return "reverse" }
func (reverseSource) SupportsReverse() bool { return true }
func (reverseSource) Open(context.Context, logsrc.Query) (logsrc.Cursor, error) {
	return nil, errors.New("not used")
}
