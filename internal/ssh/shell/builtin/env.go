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

package builtin

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// Env holds REPL I/O and callbacks for built-in command handlers.
type Env struct {
	Out       io.Writer
	In        io.Reader
	Mgr       *session.Manager
	Hist      *terminal.LineHistory
	Line      string
	ParseLine func(string) ([]string, error)
	PrintHelp func(hasActiveSession bool)
}
