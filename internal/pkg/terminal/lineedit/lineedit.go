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

// Package lineedit implements the Emacs-style raw-TTY line editor used by
// `aissh shell`, `aissh ai`, and optionally `ait` (via Options.Completer for Tab):
// ReadLine reads one edited line with history recall, Ctrl+R reverse search,
// kill keys, UTF-8 input, and Ctrl+L screen clear.
package lineedit

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

const (
	ctrlC = 3
	ctrlD = 4
	ctrlL = 12

	keyDEL            = 127
	keyBS             = '\b'
	keyESC            = 27
	csiFinalByteMin   = 0x40
	csiFinalByteMax   = 0x7e
	csiIntroducerByte = '['
)

type lineEditResult struct {
	line string
	done bool
	err  error
}

// ReadLine reads one edited line from br while stdin is in raw mode.
// The caller must put the terminal in raw mode with term.MakeRaw before calling.
// br is typically bufio.NewReader(os.Stdin); out receives echoed input and editing feedback.
// opt.Prompt is redrawn after Ctrl+L and when inserting or deleting away from the line end.
// opt.History supplies Up/Down recall when non-nil; opt.RecordHistory stores submitted lines.
// opt.Completer, when non-nil, handles Tab instead of inserting a tab (nil keeps insert-Tab).
// Returns the line without trailing newline on Enter. Empty Enter is ignored (line not submitted).
// Returns io.EOF on Ctrl+D with an empty line, or when stdin closes with no buffered input.
// Ctrl+C returns ("", nil) after writing "^C\r\n" to out. Accepts UTF-8 printable runes and Tab.
func ReadLine(br *bufio.Reader, out io.Writer, opt Options) (string, error) {
	st := newEditState()

	for {
		r, _, err := br.ReadRune()
		if err != nil {
			res := handleLineEOF(err, st.runes)
			recordSubmittedLine(opt, res.line, res.done && res.err == nil && res.line != "")
			return res.line, res.err
		}

		res := handleEditRune(br, out, opt, &st, r)
		if res.done {
			if res.line != "" && res.err == nil && opt.RecordHistory {
				recordSubmittedLine(opt, res.line, true)
			}
			return res.line, res.err
		}
	}
}

func recordSubmittedLine(opt Options, line string, submitted bool) {
	if !submitted || !opt.RecordHistory || opt.History == nil {
		return
	}
	if shouldRecordLine(line) {
		opt.History.Add(line)
	}
}

// ReadLineRaw reads one edited line without history recall or recording (legacy wrapper).
func ReadLineRaw(br *bufio.Reader, out io.Writer, prompt string) (string, error) {
	return ReadLine(br, out, Options{Prompt: prompt})
}

func handleLineEOF(err error, runes []rune) lineEditResult {
	if err == io.EOF && len(runes) > 0 {
		return lineEditResult{line: string(runes), done: true}
	}
	return lineEditResult{err: err, done: true}
}

func handleEditRune(br *bufio.Reader, out io.Writer, opt Options, st *editState, r rune) lineEditResult {
	prompt := opt.Prompt
	hist := opt.History

	switch r {
	case '\r':
		return handleLineCR(br, out, st.runes)
	case '\n':
		if len(st.runes) > 0 {
			return lineEditResult{line: string(st.runes), done: true}
		}
		return lineEditResult{}
	case '\t':
		if opt.Completer == nil {
			st.insertRune(out, prompt, '\t')
			return lineEditResult{}
		}
		applyCompleter(out, prompt, st, opt.Completer)
		return lineEditResult{}
	case ctrlC:
		_, _ = io.WriteString(out, "^C\r\n")
		return lineEditResult{line: "", done: true}
	case ctrlD:
		if len(st.runes) == 0 {
			return lineEditResult{err: io.EOF, done: true}
		}
		return lineEditResult{line: string(st.runes), done: true}
	case ctrlL:
		redrawLineAfterClear(out, prompt, st.runes, st.cursor)
		return lineEditResult{}
	case ctrlA:
		st.moveCursor(out, prompt, 0)
		return lineEditResult{}
	case ctrlE:
		st.moveCursor(out, prompt, len(st.runes))
		return lineEditResult{}
	case ctrlU:
		st.killLine(out, prompt)
		return lineEditResult{}
	case ctrlK:
		st.killToEnd(out, prompt)
		return lineEditResult{}
	case ctrlW:
		st.killWordBefore(out, prompt)
		return lineEditResult{}
	case ctrlR:
		return runReverseSearch(br, out, prompt, hist, st)
	case keyBS, keyDEL:
		st.deleteBeforeCursor(out, prompt)
		return lineEditResult{}
	case utf8.RuneError:
		return lineEditResult{}
	case keyESC:
		key, err := parseEscapeSeq(br)
		if err == nil && key != escapeNone {
			handleEscapeKey(out, prompt, st, hist, key)
		}
		return lineEditResult{}
	default:
		if !isLineInputRune(r) {
			return lineEditResult{}
		}
		st.insertRune(out, prompt, r)
		return lineEditResult{}
	}
}

func isLineInputRune(r rune) bool {
	return r == '\t' || (r != utf8.RuneError && unicode.IsPrint(r))
}

func handleLineCR(br *bufio.Reader, out io.Writer, runes []rune) lineEditResult {
	if len(runes) == 0 {
		return lineEditResult{}
	}
	_, _ = io.WriteString(out, "\r\n")

	if br.Buffered() > 0 {
		next, err := br.ReadByte()
		if err == nil && next != '\n' {
			_ = br.UnreadByte()
		}
	}
	return lineEditResult{line: string(runes), done: true}
}
