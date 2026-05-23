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

package terminal

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

const (
	ctrlC = 3
	ctrlD = 4

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

// ReadLineRaw reads one edited line from br while stdin is in raw mode.
// The caller must put the terminal in raw mode with term.MakeRaw before calling.
// br is typically bufio.NewReader(os.Stdin); out receives echoed runes and control feedback.
// Returns io.EOF on Ctrl+D with an empty line; Ctrl+C returns ("", nil) after writing "^C\r\n" to out.
// Accepts UTF-8 printable runes (including CJK); Tab (\t) is inserted for REPL line indentation.
// Backspace uses terminal display width. IME preedit is handled by the terminal, not drawn by AiT.
func ReadLineRaw(br *bufio.Reader, out io.Writer) (string, error) {
	var runes []rune
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			res := handleLineEOF(err, runes)
			return res.line, res.err
		}

		res, next := handleLineRune(br, out, runes, r)
		runes = next
		if res.done {
			return res.line, res.err
		}
	}
}

func handleLineEOF(err error, runes []rune) lineEditResult {
	if err == io.EOF && len(runes) > 0 {
		return lineEditResult{line: string(runes), done: true}
	}
	return lineEditResult{err: err, done: true}
}

func handleLineRune(br *bufio.Reader, out io.Writer, runes []rune, r rune) (lineEditResult, []rune) {
	switch r {
	case '\r':
		return handleLineCR(br, out, runes)
	case '\n':
		if len(runes) > 0 {
			return lineEditResult{line: string(runes), done: true}, runes
		}
		return lineEditResult{}, runes
	case ctrlC:
		_, _ = io.WriteString(out, "^C\r\n")
		return lineEditResult{line: "", done: true}, runes
	case ctrlD:
		if len(runes) == 0 {
			return lineEditResult{err: io.EOF, done: true}, runes
		}
		return lineEditResult{line: string(runes), done: true}, runes
	case keyBS, keyDEL:
		return lineEditResult{}, eraseLastRune(out, runes)
	case utf8.RuneError:
		return lineEditResult{}, runes
	case keyESC:
		_ = consumeEscapeSeq(br)
		return lineEditResult{}, runes
	default:
		if !isLineInputRune(r) {
			return lineEditResult{}, runes
		}
		runes = append(runes, r)
		_, _ = io.WriteString(out, string(r))
		return lineEditResult{}, runes
	}
}

func isLineInputRune(r rune) bool {
	return r == '\t' || (r != utf8.RuneError && unicode.IsPrint(r))
}

func handleLineCR(br *bufio.Reader, out io.Writer, runes []rune) (lineEditResult, []rune) {
	if len(runes) == 0 {
		return lineEditResult{}, runes
	}
	_, _ = io.WriteString(out, "\r\n")

	if br.Buffered() > 0 {
		next, err := br.ReadByte()
		if err == nil && next != '\n' {
			_ = br.UnreadByte()
		}
	}
	return lineEditResult{line: string(runes), done: true}, runes
}

func eraseLastRune(out io.Writer, runes []rune) []rune {
	if len(runes) == 0 {
		return runes
	}
	last := runes[len(runes)-1]
	runes = runes[:len(runes)-1]
	width := runewidth.RuneWidth(last)
	if width < 1 {
		width = 1
	}
	var erase strings.Builder
	for i := 0; i < width; i++ {
		erase.WriteString("\b \b")
	}
	_, _ = io.WriteString(out, erase.String())
	return runes
}

// consumeEscapeSeq reads bytes after ESC until a known sequence ends (CSI final byte csiFinalByteMin–csiFinalByteMax).
func consumeEscapeSeq(br *bufio.Reader) error {
	b, err := br.ReadByte()
	if err != nil {
		return err
	}
	if b != csiIntroducerByte {
		return nil
	}

	for {
		c, err := br.ReadByte()
		if err != nil {
			return err
		}
		if c >= csiFinalByteMin && c <= csiFinalByteMax {
			return nil
		}
	}
}
