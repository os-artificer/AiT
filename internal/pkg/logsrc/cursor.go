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

// Package logsrc reads and normalizes system logs from native tools
// (journalctl, log show, Windows Event Log) and from pure-Go built-in
// readers used when those tools are unavailable.
//
// Backends never apply a record limit: they may narrow time, unit and level
// for performance, while keyword/regex matching and the last-N cut always
// happen in Go so that -n means "last N matching records".
package logsrc

import (
	"context"
	"time"
)

// Cursor iterates records. Call Close when done; Err reports the first error
// that stopped iteration (nil on clean EOF).
type Cursor interface {
	Next(rec *Record) bool
	Err() error
	Close() error
}

// Source opens a Cursor for a Query.
//
// SupportsReverse reports whether Open (or OpenReverse for a ReverseOpener)
// yields records newest-first, which lets Collect stop after N matches
// instead of scanning the whole store.
type Source interface {
	Name() string
	SupportsReverse() bool
	Open(ctx context.Context, q Query) (Cursor, error)
}

// ReverseOpener is an optional Source capability: a separate entry point for
// newest-first iteration. Sources that encode reverse order in their own
// arguments (journalctl --reverse, EvtQueryReverseDirection) need not
// implement it and may return newest-first cursors from Open directly.
type ReverseOpener interface {
	OpenReverse(ctx context.Context, q Query) (Cursor, error)
}

// WindowedSource is an optional capability for forward-only backends whose
// store is too large to scan entirely (macOS `log show`). Open must use
// Windows()[0] (or WindowsFor(q)[0] when WindowedSourceFor is implemented);
// Collect then retries with successively larger windows until it has N
// matches or the widest window is exhausted.
type WindowedSource interface {
	Windows() []time.Duration
	OpenWindow(ctx context.Context, q Query, window time.Duration) (Cursor, error)
}

// WindowedSourceFor lets a WindowedSource choose a shorter expansion ladder
// for text-filtered queries, so rare keywords do not force multi-day
// `log show` scans that stay expensive even with --predicate.
type WindowedSourceFor interface {
	WindowsFor(q Query) []time.Duration
}
