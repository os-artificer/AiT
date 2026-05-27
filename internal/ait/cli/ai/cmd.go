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

// Package ai builds the `ait ai` Cobra subtree (chat, provider, model) and
// the local/remote AI REPL runners injected into shell and SSH packages.
package ai

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/aimode"
	"github.com/os-artificer/AiT/internal/ait/aimode/agent"
	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"

	"github.com/spf13/cobra"
)

// CLIExecutor invokes the ait root command with custom IO; cli root injects ExecuteWriters here.
type CLIExecutor func(args []string, out, errOut io.Writer) error

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr func(fn func(ctx context.Context, mgr *session.Manager) error) error

// Deps groups the cross-package dependencies the `ait ai` command needs from cli root.
type Deps struct {
	Exec       CLIExecutor
	RunWithMgr RunWithMgr
	OllamaCmd  *cobra.Command
}

// New builds the `ait ai` cobra command tree.
// d.Exec executes other ait subcommands from the AI REPL CLI tool path;
// d.RunWithMgr opens the session Manager for the chat subcommand;
// d.OllamaCmd is the `ait ai ollama` subcommand from the ollama subpackage.
func New(d Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Interactive local AI mode (AiT AI> prompt)",
		Long:  "Run AI-assisted REPL on this machine. Inside SSH use !ait ai or !! instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if helpdoc.HasHelpArg(args) {
				return helpdoc.WriteAIHelp(cmd.OutOrStdout())
			}
			modelFlag, _ := cmd.Flags().GetString("model")
			return d.RunWithMgr(func(ctx context.Context, mgr *session.Manager) error {
				return RunLocal(ctx, mgr, modelFlag, strings.Join(args, " "), d.Exec)
			})
		},
	}
	cmd.PersistentFlags().String("model", "", "Override AI provider for this session (deepseek|minimax)")
	cmd.AddCommand(newChatCmd(d), newProviderCmd(d), d.OllamaCmd)
	helpdoc.Bind(cmd, helpdoc.WriteAIHelp)
	return cmd
}

// RunLocal starts the local AI REPL with optional model override and initial line.
// modelOverride may be "" (use mgr.AI); initialLine is fed as the first REPL input.
func RunLocal(ctx context.Context, mgr *session.Manager, modelOverride, initialLine string, exec CLIExecutor) error {
	modelOverride = strings.ToLower(strings.TrimSpace(modelOverride))
	aiClient, err := localAIClient(mgr, modelOverride)
	if err != nil {
		return err
	}
	return aimode.RunREPL(ctx, localAIModeOptions(aiClient, exec), initialLine)
}

// RunLocalFromShell is the shell.SetAIModeRunner callback (no model override, empty initial line).
func RunLocalFromShell(exec CLIExecutor) func(ctx context.Context, mgr *session.Manager) error {
	return func(ctx context.Context, mgr *session.Manager) error {
		return RunLocal(ctx, mgr, "", "", exec)
	}
}

// RemoteRunner builds the sshclient.RemoteAIRunner injected into Manager for SSH AI mode.
// exec wires the in-REPL `ait` CLI dispatch so users can run local subcommands from remote chat.
func RemoteRunner(exec CLIExecutor) sshclient.RemoteAIRunner {
	return func(p sshclient.RemoteAIParams) error {
		return runRemoteAIMode(p, exec)
	}
}

func localAIClient(mgr *session.Manager, modelOverride string) (*ai.Client, error) {
	if modelOverride == "" && mgr.AI != nil {
		return mgr.AI, nil
	}
	var opts []ai.ClientOption
	if modelOverride != "" {
		opts = append(opts, ai.WithProviderOverride(modelOverride))
	}
	return ai.NewClient(mgr.Config, opts...)
}

func localAIModeOptions(aiClient *ai.Client, exec CLIExecutor) aimode.Options {
	return aimode.Options{
		Scope:   ai.ScopeLocal,
		AI:      aiClient,
		Local:   aimode.DefaultLocalRunner(),
		Capture: agent.DefaultLocalCaptureRunner(),
		CLI:     func(argv []string, out, errOut io.Writer) error { return exec(argv, out, errOut) },
		In:      os.Stdin,
		Out:     os.Stdout,
		ErrOut:  os.Stderr,
	}
}

type ptyRemoteRunner struct {
	stdin io.Writer
}

func (r *ptyRemoteRunner) Run(cmd string) error {
	if r.stdin == nil {
		return fmt.Errorf("remote stdin not available")
	}
	_, err := io.WriteString(r.stdin, cmd+"\n")
	return err
}

func runRemoteAIMode(p sshclient.RemoteAIParams, exec CLIExecutor) error {
	if p.AI == nil {
		return ai.NotConfiguredErr()
	}

	var in io.Reader = os.Stdin
	useTTY := false
	if p.ReplIn != nil {
		in = p.ReplIn
		useTTY = true
	}

	opts := aimode.Options{
		Scope:          ai.ScopeRemote,
		AI:             p.AI,
		HostLabel:      p.HostLabel,
		RemoteCwd:      p.RemoteCwd,
		Scrollback:     p.Scrollback,
		Remote:         &ptyRemoteRunner{stdin: p.Stdin},
		CLI:            func(argv []string, out, errOut io.Writer) error { return exec(argv, out, errOut) },
		In:             in,
		Out:            os.Stdout,
		ErrOut:         os.Stderr,
		UseTTYLineEdit: useTTY,
	}
	if p.RunCapture != nil {
		opts.Capture = agent.NewCaptureRunner(func(ctx context.Context, cmd string) (string, error) {
			_ = ctx
			return p.RunCapture(cmd)
		})
	}
	return aimode.RunREPL(p.Ctx, opts, p.InitialQuestion)
}
