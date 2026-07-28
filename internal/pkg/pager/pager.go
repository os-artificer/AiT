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

// Package pager displays line-oriented output one screen at a time, with
// less-like keys, on an interactive terminal. Non-terminal writers and
// explicitly disabled paging copy every line straight through, so piped and
// redirected output is unchanged.
package pager

import (
	"context"
	"io"
)

const (
	// DefaultMaxLines caps how many lines interactive mode buffers.
	DefaultMaxLines = 200_000
	// HintLoading marks a source that is still producing lines.
	HintLoading = "loading"
	// HintScanning marks a bounded scan that must finish before display.
	HintScanning = "scanning"
	// fallbackRows/fallbackCols are used when the terminal size is unknown.
	fallbackRows = 24
	fallbackCols = 80
)

// Options configures Page.
type Options struct {
	// Disabled copies lines through without paging.
	Disabled bool
	// Bounded waits for the source to finish before drawing the first frame;
	// unbounded sources draw as soon as one screen is available.
	Bounded bool
	// MaxLines overrides DefaultMaxLines.
	MaxLines int
	// Rows and Cols override the detected terminal size.
	Rows, Cols int
	// StatusHint labels the progress marker (HintLoading or HintScanning).
	StatusHint string
	// Input overrides the key source; defaults to os.Stdin.
	Input io.Reader
}

// LineSource yields the lines to display. NextLine blocks until a line is
// available and returns false when the source is exhausted; Err then reports
// why iteration stopped, or nil for a clean end.
type LineSource interface {
	NextLine() (string, bool)
	Err() error
}

// Page displays src on w. It pages interactively when w is a terminal and
// opt.Disabled is false, and otherwise writes every line to w.
//
// Returns the source's error, a write error, or nil when the user quit. A
// cancelled ctx ends paging without an error.
func Page(ctx context.Context, w io.Writer, src LineSource, opt Options) error {
	if opt.MaxLines <= 0 {
		opt.MaxLines = DefaultMaxLines
	}
	fd, isTTY := writerTerminal(w)
	if opt.Disabled || !isTTY {
		return copyAll(ctx, w, src)
	}
	return pageInteractive(ctx, w, fd, src, opt)
}
