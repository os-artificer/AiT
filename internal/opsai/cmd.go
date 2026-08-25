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

package opsai

import (
	"context"

	"github.com/os-artificer/AiT/internal/opsai/connect"
	"github.com/os-artificer/AiT/internal/opsai/deps"
	"github.com/os-artificer/AiT/internal/opsai/subregistry"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Deps is an alias for shared ai subtree dependencies.
type Deps = deps.Deps

// CLIExecutor invokes aissh (or toolbox CLIs) with custom IO.
type CLIExecutor = deps.CLIExecutor

// RunWithMgr opens config and AI client under a signal-cancellable context.
type RunWithMgr = deps.RunWithMgr

// New builds the root `ai` cobra command tree.
func New(d Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           toolname.AI,
		Short:         "Ops AI Agent (experimental; off by default)",
		Long:          pkghelpdoc.LongHint(toolname.AI),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if skipAIExperimentalRequire(cmd, args) {
				return nil
			}
			return config.RequireAIExperimental(nil)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteAIHelp(cmd.OutOrStdout())
			}
			if len(args) == 0 {
				if err := WriteAIHelp(cmd.OutOrStdout()); err != nil {
					return err
				}
				return errhint.WithTryMsg(
					"interactive agent loop moved to ait agent",
					"Use the ait toolbox agent loop for multi-turn tasks.",
					clihints.TryAitAgent,
				)
			}
			if len(args) > 0 && connect.LooksLikeDestination(args[0]) {
				return d.RunWithMgr(func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error {
					return connect.Run(ctx, cfg, aiClient, args, d.Exec)
				})
			}
			return errhint.WithTryMsg(
				"unknown ai command",
				"Use ai subcommands or ait agent for the task loop.",
				toolname.AICLI("ask ping"),
				clihints.TryAitAgent,
			)
		},
	}
	cmd.PersistentFlags().String("model", "", "Override AI provider for this session (deepseek|minimax)")
	subregistry.Attach(cmd, deps.Deps{Exec: d.Exec, RunWithMgr: d.RunWithMgr, ApplyOpsDomain: d.ApplyOpsDomain})
	pkghelpdoc.Bind(cmd, WriteAIHelp)
	return cmd
}

// RunLocal is deprecated; bare ai prints help. Kept for tests that import the symbol.
func RunLocal(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client, modelOverride, initialLine string, exec CLIExecutor) error {
	_ = ctx
	_ = cfg
	_ = aiClient
	_ = modelOverride
	_ = initialLine
	_ = exec
	return errhint.WithTryMsg(
		"interactive agent loop moved to ait agent",
		"Use the ait toolbox agent loop for multi-turn tasks.",
		clihints.TryAitAgent,
	)
}
