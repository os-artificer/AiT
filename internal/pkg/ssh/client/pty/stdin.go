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

// Package pty hosts the PTY-side helpers used by an interactive SSH session:
// the single-goroutine stdin Reader that routes bytes between SSH forwarding
// and the local AI REPL, a one-byte-per-channel ReplReader, and the cross-
// platform terminal resize watcher. Receiver-bearing types live here; SSH
// session wiring (Client.RunInteractive, *Client method receivers) stays in
// the parent ssh/client package.
package pty

import (
	"io"
	"os"
	"sync/atomic"
)

const readerBufSize = 32

// StdinByte is one byte plus a sticky error from the underlying reader.
type StdinByte struct {
	B   byte
	Err error
}

// Reader is the sole goroutine that reads the interactive TTY stdin stream.
// During remote AI REPL bytes are routed to ReplCh instead of the SSH
// forward channel.
type Reader struct {
	Ch         chan StdinByte
	stopCh     chan struct{}
	replCh     chan byte
	replActive atomic.Bool
}

// Start reads os.Stdin in a goroutine and returns the controller.
func Start() *Reader {
	return StartFrom(os.Stdin)
}

// StartFrom is like Start but reads from r (tests only).
func StartFrom(r io.Reader) *Reader {
	sr := &Reader{
		Ch:     make(chan StdinByte, readerBufSize),
		stopCh: make(chan struct{}),
	}

	go sr.readLoop(r)
	return sr
}

// BeginREPL routes subsequent stdin bytes to the returned io.Reader until EndREPL.
// The caller's SSH forward loop must be blocked while the REPL runs.
func (sr *Reader) BeginREPL() io.Reader {
	sr.drainForward()
	sr.replCh = make(chan byte, readerBufSize)
	sr.replActive.Store(true)
	return NewReplReader(sr.replCh)
}

// DrainForward discards bytes queued for SSH forwarding (tests and ZMODEM prep).
func (sr *Reader) DrainForward() {
	sr.drainForward()
}

// REPLActive reports whether stdin is routed to the AI REPL instead of SSH forwarding.
func (sr *Reader) REPLActive() bool {
	return sr.replActive.Load()
}

// EndREPL stops routing stdin to the AI REPL and discards unread repl bytes.
func (sr *Reader) EndREPL() {
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

// Stop asks the reader goroutine to exit.
func (sr *Reader) Stop() {
	close(sr.stopCh)
}

func (sr *Reader) readLoop(r io.Reader) {
	defer close(sr.Ch)

	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if sr.replActive.Load() {
				if !sr.sendRepl(buf[0]) {
					return
				}
			} else if !sr.sendForward(StdinByte{B: buf[0]}) {
				return
			}
		}
		if err != nil {
			if !sr.replActive.Load() {
				sr.sendForward(StdinByte{Err: err})
			}
			return
		}
	}
}

func (sr *Reader) sendForward(sb StdinByte) bool {
	select {
	case sr.Ch <- sb:
		return true
	case <-sr.stopCh:
		return false
	}
}

func (sr *Reader) sendRepl(b byte) bool {
	select {
	case sr.replCh <- b:
		return true
	case <-sr.stopCh:
		return false
	}
}

func (sr *Reader) drainForward() {
	for {
		select {
		case <-sr.Ch:
		default:
			return
		}
	}
}
