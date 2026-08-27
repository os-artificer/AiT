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

// Package procsig holds process signal helpers shared by REPL and agent loops
// that spawn child processes.
package procsig

import (
	"os"
	"os/signal"
)

// WithAbsorbedInterrupt runs fn while the parent process absorbs os.Interrupt
// (Ctrl+C) via Notify+drain, so a REPL or agent loop survives a Ctrl+C aimed at
// the child process started inside fn.
// Child processes started inside fn still receive the default SIGINT
// disposition after exec; signal.Ignore must not be used here because Ignore
// before exec is inherited as SIG_IGN.
// fn must not be nil. Returns the error returned by fn.
func WithAbsorbedInterrupt(fn func() error) error {
	if fn == nil {
		return nil
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer signal.Stop(ch)

	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-ch:
			case <-done:
				return
			}
		}
	}()

	return fn()
}
