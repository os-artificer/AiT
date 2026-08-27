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

package connect

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/opsai/deps"
	"github.com/os-artificer/AiT/internal/opsai/runner"
	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	acrunner "github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/agentui/plain"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// reservedTokens never route as SSH destinations (OA17).
var reservedTokens = map[string]struct{}{
	"ask": {}, "agent": {}, "connect": {}, "version": {},
	"provider": {}, "ollama": {}, "model": {}, "help": {}, "-h": {}, "--help": {},
	// Loop meta verbs (must not bare-route to connect).
	"mode": {}, "approve": {}, "trust": {}, "policy": {}, "run": {},
	"history": {}, "clear": {}, "exit": {}, "quit": {}, "disconnect": {},
	toolname.AISsh: {},
}

// LooksLikeDestination reports whether first may be an SSH destination.
func LooksLikeDestination(first string) bool {
	first = strings.TrimSpace(first)
	if first == "" {
		return false
	}
	if _, reserved := reservedTokens[strings.ToLower(first)]; reserved {
		return false
	}
	if strings.HasPrefix(first, "-") {
		return false
	}
	return true
}

// Run dials destination argv and runs a remote Ops Agent session (no aissh meta).
func Run(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client, argv []string, exec deps.CLIExecutor) error {
	if err := config.RequireAIExperimental(cfg); err != nil {
		return err
	}
	if len(argv) == 0 {
		return errhint.WithTryMsg("connect requires a destination",
			"Pass an SSH destination after connect (or as the first ai argument).",
			toolname.AICLI("connect user@host"),
			toolname.AICLI("user@host"),
		)
	}

	spec, err := openssh.ParseLogin(argv)
	if err != nil {
		return err
	}
	if spec.Tmux {
		return errhint.WithTryMsg("--tmux is not supported for ai connect",
			"ai connect opens an Agent session without tmux keep-alive.",
			toolname.AICLI("connect user@host"),
		)
	}
	if len(spec.RemoteCommand) > 0 {
		return errhint.WithTryMsg("remote command is not supported for ai connect",
			"ai connect runs an interactive Agent session only.",
			toolname.AICLI("connect user@host"),
		)
	}

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return err
	}

	client, err := sshclient.DialChain(resolved.ClientOptions())
	if err != nil {
		return err
	}
	defer client.Close()

	limits := agentLimits{}
	if cfg != nil {
		limits.maxSteps = cfg.AgentMaxSteps()
		limits.commandTimeout = cfg.AgentCommandTimeout()
	}
	if aiClient == nil {
		aiClient, err = runner.NewAIClient(cfg, "")
		if err != nil {
			return err
		}
	}

	return client.RunAgentSession(ctx, sshclient.AgentSessionOpts{
		HostLabel: resolved.User + "@" + resolved.Host,
		OnReady: func(ctx context.Context, h sshclient.AgentSessionHandles) error {
			return runRemoteAgentLoop(ctx, aiClient, h, limits, exec)
		},
	})
}

// agentLimits bounds one Agent session: turns per goal and wall clock per tool
// command. A zero field falls back to the agentcore default.
type agentLimits struct {
	maxSteps       int
	commandTimeout time.Duration
}

// buildRemoteAgentConfig assembles agentcore.Config for a remote Agent session (OA5/OA6).
func buildRemoteAgentConfig(aiClient *pkgai.Client, h sshclient.AgentSessionHandles, limits agentLimits, exec deps.CLIExecutor) agentcore.Config {
	out := os.Stdout
	renderer := plain.Renderer{Out: out}
	var remote agentcore.RemoteRunner
	if h.Stdin != nil {
		remote = &ptyRemoteRunner{stdin: h.Stdin}
	}
	return agentcore.Config{
		Scope:      pkgai.ScopeRemote,
		AI:         aiClient,
		HostLabel:  h.HostLabel,
		Scrollback: h.Scrollback,
		Remote:     remote,
		Capture: acrunner.NewCaptureRunner(func(_ context.Context, cmd string) (string, error) {
			return h.RunCapture(cmd)
		}),
		CLI: func(argv []string, o, e io.Writer) error {
			return exec(argv, o, e)
		},
		Toolbox:        dispatchToolbox,
		In:             os.Stdin,
		Out:            out,
		ErrOut:         os.Stderr,
		MaxSteps:       limits.maxSteps,
		CommandTimeout: limits.commandTimeout,
		Domain:         "ops",
		Emitter:        renderer.NewEmitter(),
	}
}

type ptyRemoteRunner struct {
	stdin io.Writer
}

func (r *ptyRemoteRunner) Run(cmd string) error {
	if r.stdin == nil {
		return errhint.WithTryMsg("remote stdin not available", "", toolname.AICLI("connect user@host"))
	}
	_, err := io.WriteString(r.stdin, cmd+"\n")
	return err
}

func runRemoteAgentLoop(ctx context.Context, aiClient *pkgai.Client, h sshclient.AgentSessionHandles, limits agentLimits, exec deps.CLIExecutor) error {
	if aiClient == nil {
		return pkgai.NotConfiguredErr()
	}
	cfg := buildRemoteAgentConfig(aiClient, h, limits, exec)
	return agentcore.RunLoop(ctx, cfg, "")
}

// dispatchToolbox is set from opsai to avoid import cycles; defaults to PATH LookPath of known bins.
var dispatchToolbox = defaultDispatchToolbox

func defaultDispatchToolbox(argv []string, out, errOut io.Writer) error {
	return errhint.WithTryMsg("toolbox not configured",
		"Remote Agent toolbox requires the ai binary DispatchToolbox.",
		toolname.AICLI(`agent "check disk"`),
	)
}

// SetToolboxDispatch wires agentops.DispatchToolbox (called from opsai init).
func SetToolboxDispatch(fn func([]string, io.Writer, io.Writer) error) {
	if fn != nil {
		dispatchToolbox = fn
	}
}
