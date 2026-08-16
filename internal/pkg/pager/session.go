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

package pager

import (
	"context"
	"io"
)

// session owns one interactive paging run: the line buffer, the incoming line
// channel, the key reader and the status line.
type session struct {
	out        io.Writer
	src        LineSource
	lines      <-chan string
	keys       *keyReader
	vp         *Viewport
	cols       int
	viewRows   int
	maxLines   int
	bounded    bool
	hint       string
	done       bool
	truncated  bool
	searchTerm string
	searchBack bool
	err        error
	stopFeed   context.CancelFunc
}

// newSession wires a paging session. One row is reserved for the status line.
func newSession(out io.Writer, src LineSource, lines <-chan string, keys *keyReader, rows, cols int, opt Options) *session {
	viewRows := rows - 1
	if viewRows < 1 {
		viewRows = 1
	}
	return &session{
		out:      out,
		src:      src,
		lines:    lines,
		keys:     keys,
		vp:       NewViewport(viewRows),
		cols:     cols,
		viewRows: viewRows,
		maxLines: opt.MaxLines,
		bounded:  opt.Bounded,
		hint:     startHint(opt),
	}
}

// run fills the first screen, then serves keys and incoming lines until the
// user quits, the terminal closes, or ctx is cancelled.
func (s *session) run(ctx context.Context) error {
	if quit := s.prefill(ctx); quit {
		return s.finish()
	}
	s.redraw()

	for {
		select {
		case <-ctx.Done():
			return s.finish()
		case line, ok := <-s.lines:
			s.consume(line, ok, true)
		case key, ok := <-s.keys.stream():
			if !ok {
				return s.finish()
			}
			if s.handleKey(key) {
				return s.finish()
			}
			s.redraw()
		}
	}
}

// prefill buffers the lines needed before the first frame: everything for a
// bounded source, one screen for a progressive one. It reports whether the
// user asked to quit while waiting.
func (s *session) prefill(ctx context.Context) bool {
	for s.lines != nil {
		if !s.bounded && s.vp.Len() >= s.viewRows {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case line, ok := <-s.lines:
			s.consume(line, ok, false)
		case key, ok := <-s.keys.stream():
			if !ok || isQuitKey(key) {
				return true
			}
		}
	}
	return false
}

// consume appends one line, or records that the source ended. draw is false
// during prefill, where no frame has been painted yet.
func (s *session) consume(line string, ok, draw bool) {
	if !ok {
		s.lines = nil
		s.done = true
		s.hint = ""
		s.err = s.src.Err()
		s.drawStatusIf(draw)
		return
	}
	if s.vp.Len() >= s.maxLines {
		s.lines = nil
		s.done = true
		s.truncated = true
		s.hint = ""
		s.cancelFeed()
		s.drawStatusIf(draw)
		return
	}

	s.vp.Append(line)
	if !draw {
		return
	}
	// Only the still-filling first screen shifts content; later arrivals just
	// move the progress indicator.
	if s.vp.Len() <= s.viewRows {
		s.redraw()
		return
	}
	s.drawStatus()
}

// handleKey applies one keystroke and reports whether to quit.
func (s *session) handleKey(key byte) bool {
	switch key {
	case 'q', 'Q', keyCtrlC:
		return true
	case ' ', 'f':
		s.vp.Down(s.viewRows)
	case 'b':
		s.vp.Up(s.viewRows)
	case 'j', '\n', '\r':
		s.vp.Down(1)
	case 'k':
		s.vp.Up(1)
	case 'd':
		s.vp.HalfDown()
	case 'u':
		s.vp.HalfUp()
	case 'g':
		s.vp.Top()
	case 'G':
		s.vp.Bottom()
	case 'h', '?':
		s.showHelp()
	case '/':
		s.startSearch(false)
	case '\\':
		s.startSearch(true)
	case 'n':
		s.repeatSearch(false)
	case 'N':
		s.repeatSearch(true)
	case keyEscape:
		s.keys.readArrow(s.vp)
	}
	return false
}

func (s *session) showHelp() {
	writeHelp(s.out)
	s.keys.discardOne()
}

// startSearch prompts for a term and jumps to the first match. backward
// searches towards the start of the buffer.
func (s *session) startSearch(backward bool) {
	prompt := searchPromptForward
	if backward {
		prompt = searchPromptBackward
	}
	query := s.keys.readLine(s.out, s.cols, prompt)
	if query == "" {
		return
	}
	s.searchTerm = query
	s.searchBack = backward
	s.vp.Search(query, backward)
}

func (s *session) repeatSearch(invert bool) {
	if s.searchTerm == "" {
		return
	}
	backward := s.searchBack
	if invert {
		backward = !backward
	}
	s.vp.Search(s.searchTerm, backward)
}

// redraw repaints the visible lines and the status line.
func (s *session) redraw() {
	writeFrame(s.out, s.vp.Visible())
	s.drawStatus()
}

func (s *session) drawStatus() {
	writeStatus(s.out, formatStatus(s.vp.Percent(), s.done, s.truncated, s.hint), s.cols)
}

func (s *session) drawStatusIf(draw bool) {
	if draw {
		s.drawStatus()
	}
}

// finish leaves the alternate cursor position on a fresh line and reports the
// source error, if any.
func (s *session) finish() error {
	s.cancelFeed()
	writeNewline(s.out)
	return s.err
}

func (s *session) cancelFeed() {
	if s.stopFeed != nil {
		s.stopFeed()
		s.stopFeed = nil
	}
}

func startHint(opt Options) string {
	if opt.StatusHint != "" {
		return opt.StatusHint
	}
	if opt.Bounded {
		return HintScanning
	}
	return HintLoading
}
