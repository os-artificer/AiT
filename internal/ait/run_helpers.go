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

package ait

import (
	"os"
	"os/signal"
)

// withAbsorbedInterrupt runs fn while the parent process absorbs os.Interrupt
// (Ctrl+C) via Notify+drain so the REPL does not exit. Child processes started
// inside fn still receive default SIGINT disposition after exec (AT29).
// Do not use signal.Ignore here: Ignore before exec is inherited as SIG_IGN.
func withAbsorbedInterrupt(fn func() error) error {
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
