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

// Package ai builds the `ai-ssh ai` Cobra subtree (chat, provider, model) and
// the local/remote AI REPL runners injected into shell and SSH packages.
package ai

import (
	"context"
	"strings"

	"github.com/os-artificer/AiT/internal/aissh/ai/deps"
	"github.com/os-artificer/AiT/internal/aissh/ai/mode"
	"github.com/os-artificer/AiT/internal/aissh/ai/subregistry"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"

	"github.com/spf13/cobra"
)

// Deps is an alias for shared ai subtree dependencies.
type Deps = deps.Deps

// CLIExecutor invokes the ai-ssh root command with custom IO; cli root injects ExecuteWriters here.
type CLIExecutor = deps.CLIExecutor

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr = deps.RunWithMgr

// New builds the `ai-ssh ai` cobra command tree.
// d.Exec executes other ai-ssh subcommands from the AI REPL CLI tool path;
// d.RunWithMgr opens the session Manager for the chat subcommand.
func New(d Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           Name,
		Short:         "Interactive local AI mode (AiT AI(Agent)> prompt)",
		Long:          pkghelpdoc.LongHint("ai"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteAIHelp(cmd.OutOrStdout())
			}
			modelFlag, _ := cmd.Flags().GetString("model")
			return d.RunWithMgr(func(ctx context.Context, mgr *session.Manager) error {
				return RunLocal(ctx, mgr, modelFlag, strings.Join(args, " "), d.Exec)
			})
		},
	}
	cmd.PersistentFlags().String("model", "", "Override AI provider for this session (deepseek|minimax)")
	subregistry.Attach(cmd, deps.Deps{Exec: d.Exec, RunWithMgr: d.RunWithMgr})
	pkghelpdoc.Bind(cmd, WriteAIHelp)
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
	return mode.RunREPL(ctx, localAIModeOptions(aiClient, mgr.Config.AgentMaxSteps(), exec), initialLine)
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
