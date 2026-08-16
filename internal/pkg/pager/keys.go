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
	"bufio"
	"io"
	"strings"
)

const (
	keyCtrlC  = 0x03
	keyEscape = 0x1b
	// Terminals send either DEL or BS for the backspace key.
	keyDEL       = 0x7f
	keyBS        = 0x08
	keyCSI       = '['
	keyArrowUp   = 'A'
	keyArrowDown = 'B'

	searchPromptForward  = "/"
	searchPromptBackward = "?"

	// keyBuffer holds keystrokes typed while a frame is being painted.
	keyBuffer = 64
)

// keyReader turns a byte stream into a channel of keystrokes so that the
// paging loop can wait on input and incoming lines at the same time. A raw
// terminal delivers Ctrl+C as a byte rather than a signal, so quitting is
// handled here as an ordinary key.
type keyReader struct {
	ch <-chan byte
}

// newKeyReader starts reading r in the background. The channel closes when r
// reaches EOF or errors; the goroutine ends with it.
func newKeyReader(r io.Reader) *keyReader {
	ch := make(chan byte, keyBuffer)
	go func() {
		defer close(ch)
		br := bufio.NewReader(r)
		for {
			b, err := br.ReadByte()
			if err != nil {
				return
			}
			ch <- b
		}
	}()
	return &keyReader{ch: ch}
}

// stream exposes the keystroke channel for use in a select.
func (k *keyReader) stream() <-chan byte { return k.ch }

// next blocks for one keystroke. The second result is false once the input
// stream has ended.
func (k *keyReader) next() (byte, bool) {
	b, ok := <-k.ch
	return b, ok
}

// discardOne waits for any keystroke, used to dismiss the help screen.
func (k *keyReader) discardOne() { k.next() }

// readLine reads a search term on the status line, echoing it after prompt.
// Backspace edits, Enter accepts, and Ctrl+C or Escape cancels and returns "".
func (k *keyReader) readLine(out io.Writer, cols int, prompt string) string {
	var b strings.Builder
	writeStatus(out, prompt, cols)

	for {
		ch, ok := k.next()
		if !ok {
			return b.String()
		}
		switch ch {
		case '\n', '\r':
			return b.String()
		case keyCtrlC, keyEscape:
			return ""
		case keyDEL, keyBS:
			trimLastRune(&b)
		default:
			b.WriteByte(ch)
		}
		writeStatus(out, prompt+b.String(), cols)
	}
}

// readArrow consumes the remainder of a CSI escape sequence and applies Up or
// Down to vp. Unknown sequences are ignored.
func (k *keyReader) readArrow(vp *Viewport) {
	if b, ok := k.next(); !ok || b != keyCSI {
		return
	}
	final, ok := k.next()
	if !ok {
		return
	}
	switch final {
	case keyArrowUp:
		vp.Up(1)
	case keyArrowDown:
		vp.Down(1)
	}
}

func isQuitKey(b byte) bool {
	return b == 'q' || b == 'Q' || b == keyCtrlC
}

// trimLastRune removes the final rune from b, keeping multi-byte input intact.
func trimLastRune(b *strings.Builder) {
	s := b.String()
	if s == "" {
		return
	}
	runes := []rune(s)
	b.Reset()
	b.WriteString(string(runes[:len(runes)-1]))
}
