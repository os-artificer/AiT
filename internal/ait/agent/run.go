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

// Package agent wires ait agent / ait> agent into agentcore + plain UI.
package agent

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	"github.com/os-artificer/AiT/internal/pkg/agentops"
	"github.com/os-artificer/AiT/internal/pkg/agentui/plain"
	"github.com/os-artificer/AiT/internal/pkg/agentui/tui"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"golang.org/x/term"
)

// Options configures an ait agent session.
type Options struct {
	Nested bool
	Goal   string
	Plain  bool
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// RunLoop starts the v4 agent task loop with experimental gate and plain UI.
func RunLoop(ctx context.Context, opt Options) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := config.RequireAIExperimental(cfg); err != nil {
		return err
	}

	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()

	aiClient, err := pkgai.NewClient(cfg)
	if err != nil {
		return pkgai.NotConfiguredErr()
	}

	in, out, errOut := opt.In, opt.Out, opt.ErrOut
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	renderer := plain.Renderer{Out: out}
	warn := func(msg string) {
		renderer.NewEmitter().Emit(agentcore.Event{Kind: agentcore.EventWarning, Detail: msg})
	}
	envExtra := agentops.BuildContext(ctx, warn)

	maxSteps := 0
	if cfg != nil {
		maxSteps = cfg.AgentMaxSteps()
	}

	acfg := agentcore.Config{
		Scope:    pkgai.ScopeLocal,
		AI:       aiClient,
		CLI:      execAisshCLI,
		Toolbox:  agentops.DispatchToolbox,
		In:       in,
		Out:      out,
		ErrOut:   errOut,
		MaxSteps: maxSteps,
		Domain:   "ops",
		Emitter:  renderer.NewEmitter(),
		Nested:   opt.Nested,
		EnvExtra: envExtra,
	}

	goal := opt.Goal
	if goal == "" && !isTTY(in) {
		return agentcore.RunLoop(ctx, acfg, "")
	}
	if isTTY(in) && !opt.Plain {
		outFile, outOK := out.(*os.File)
		if outOK && term.IsTerminal(int(outFile.Fd())) {
			return tui.RunLoop(ctx, acfg, goal)
		}
	}
	return agentcore.RunLoop(ctx, acfg, goal)
}

func execAisshCLI(args []string, out, errOut io.Writer) error {
	bin, err := exec.LookPath(toolname.AISsh)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func isTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
