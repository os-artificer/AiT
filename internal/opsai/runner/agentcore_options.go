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

package runner

import (
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/agentops"
	"github.com/os-artificer/AiT/internal/pkg/agentui/plain"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
)

// LocalAgentConfig builds agentcore.Config for local ops agent sessions.
func LocalAgentConfig(aiClient *pkgai.Client, maxSteps int, exec func([]string, io.Writer, io.Writer) error) agentcore.Config {
	out := os.Stdout
	renderer := plain.Renderer{Out: out}
	return agentcore.Config{
		Scope:    pkgai.ScopeLocal,
		AI:       aiClient,
		Capture:  runner.DefaultLocalCaptureRunner(),
		CLI:      func(argv []string, o, e io.Writer) error { return exec(argv, o, e) },
		Toolbox:  agentops.DispatchToolbox,
		In:       os.Stdin,
		Out:      out,
		ErrOut:   os.Stderr,
		MaxSteps: maxSteps,
		Domain:   "ops",
		Emitter:  renderer.NewEmitter(),
	}
}
