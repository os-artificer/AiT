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
	"context"
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/ait/agent"
	"github.com/os-artificer/AiT/internal/pkg/agentcore"

	"golang.org/x/term"
)

// runNestedAgent runs the v4 agent loop from ait> with TTY restore when stdin is a terminal.
func runNestedAgent(ctx context.Context, goal string, out, errOut io.Writer) error {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		old, err := term.MakeRaw(fd)
		if err == nil {
			defer func() { _ = term.Restore(fd, old) }()
			if restoreErr := term.Restore(fd, old); restoreErr != nil {
				return fmt.Errorf("terminal: %w", restoreErr)
			}
			runErr := agent.RunLoop(ctx, agent.Options{
				Nested: true,
				Goal:   goal,
				In:     os.Stdin,
				Out:    out,
				ErrOut: errOut,
			})
			if _, rawErr := term.MakeRaw(fd); rawErr != nil {
				if runErr != nil {
					return fmt.Errorf("terminal: %w", rawErr)
				}
				return fmt.Errorf("terminal: %w", rawErr)
			}
			return runErr
		}
	}
	return agent.RunLoop(ctx, agent.Options{
		Nested: true,
		Goal:   goal,
		In:     os.Stdin,
		Out:    out,
		ErrOut: errOut,
	})
}

// agentLoopEnded reports whether err is a normal nested agent exit.
func agentLoopEnded(err error) bool {
	return err == nil || err == agentcore.ErrSessionEnded
}
