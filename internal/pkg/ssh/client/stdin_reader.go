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

package client

import (
	"io"
	"os"
	"sync/atomic"
)

const stdinReaderBufSize = 32

// stdinReader is the sole goroutine that reads the interactive TTY stdin stream.
// During remote AI REPL, bytes route to replCh instead of the SSH forward channel.
type stdinReader struct {
	ch         chan stdinByte
	stopCh     chan struct{}
	replCh     chan byte
	replActive atomic.Bool
}

// startStdinReader starts reading os.Stdin and returns a controller for REPL routing.
func startStdinReader() *stdinReader {
	return startStdinReaderFrom(os.Stdin)
}

// startStdinReaderFrom is like startStdinReader but reads from r (tests only).
func startStdinReaderFrom(r io.Reader) *stdinReader {
	sr := &stdinReader{
		ch:     make(chan stdinByte, stdinReaderBufSize),
		stopCh: make(chan struct{}),
	}

	go sr.readLoop(r)
	return sr
}

func (sr *stdinReader) readLoop(r io.Reader) {
	defer close(sr.ch)

	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if sr.replActive.Load() {
				if !sr.sendRepl(buf[0]) {
					return
				}
			} else if !sr.sendForward(stdinByte{b: buf[0]}) {
				return
			}
		}
		if err != nil {
			if !sr.replActive.Load() {
				sr.sendForward(stdinByte{err: err})
			}
			return
		}
	}
}

func (sr *stdinReader) sendForward(sb stdinByte) bool {
	select {
	case sr.ch <- sb:
		return true
	case <-sr.stopCh:
		return false
	}
}

func (sr *stdinReader) sendRepl(b byte) bool {
	select {
	case sr.replCh <- b:
		return true
	case <-sr.stopCh:
		return false
	}
}

// beginREPL routes subsequent stdin bytes to replIn until endREPL.
// forwardStdin must be blocked (e.g. inside handleAIMetaLine) while the REPL runs.
func (sr *stdinReader) beginREPL() io.Reader {
	sr.drainForward()
	sr.replCh = make(chan byte, stdinReaderBufSize)
	sr.replActive.Store(true)
	return newReplStdinReader(sr.replCh)
}

// endREPL stops routing stdin to the AI REPL and discards unread repl bytes.
func (sr *stdinReader) endREPL() {
	sr.replActive.Store(false)
	if sr.replCh != nil {
		for {
			select {
			case <-sr.replCh:
			default:
				sr.replCh = nil
				return
			}
		}
	}
}

// drainForward drops bytes already queued for SSH forwarding.
func (sr *stdinReader) drainForward() {
	for {
		select {
		case <-sr.ch:
		default:
			return
		}
	}
}

// stop asks the reader goroutine to exit.
func (sr *stdinReader) stop() {
	close(sr.stopCh)
}
