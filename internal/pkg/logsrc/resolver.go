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

package logsrc

import (
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Resolver holds the injectable dependencies used while probing and opening
// sources. Tests substitute LookPath to prove that a built-in backend never
// consults PATH, and Notice to capture fallback messages.
type Resolver struct {
	LookPath func(string) (string, error)
	Notice   io.Writer
}

// DefaultResolver returns a production Resolver using platform.LookPath and
// os.Stderr for notices.
func DefaultResolver() *Resolver {
	return &Resolver{
		LookPath: platform.LookPath,
		Notice:   os.Stderr,
	}
}

func (r *Resolver) lookPath(name string) (string, error) {
	if r == nil || r.LookPath == nil {
		return platform.LookPath(name)
	}
	return r.LookPath(name)
}

// toolPath resolves the executable for a native source. override, when set,
// wins over PATH so that tests can point at a stub. Returns an ErrToolMissing
// error naming tool when it is not on PATH, which makes the resolver fall
// through to the next layer in the chain.
func (r *Resolver) toolPath(override, tool string) (string, error) {
	if override != "" {
		return override, nil
	}
	path, err := r.lookPath(tool)
	if err != nil {
		return "", fmtErrTool(tool)
	}
	return path, nil
}

func (r *Resolver) notice() io.Writer {
	if r == nil || r.Notice == nil {
		return io.Discard
	}
	return r.Notice
}

// WriteNotice writes one tool-prefixed advisory line to w unless quiet is set
// or w discards output. Notices are informational: they never carry errors.
func WriteNotice(w io.Writer, quiet bool, format string, args ...any) {
	if quiet || w == nil || w == io.Discard {
		return
	}
	_, _ = fmt.Fprintf(w, toolname.AILog+": "+format+"\n", args...)
}
