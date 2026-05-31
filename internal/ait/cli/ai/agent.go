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

package ai

import (
	"context"
	"os"

	"github.com/os-artificer/AiT/internal/ait/aimode"
	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newAgentCmd(d Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [message...]",
		Short: "Send one message to the AI in agent mode and execute tool calls (non-interactive)",
		Long:  "Run 'ait ai agent --help' for usage. With no arguments, reads the prompt from stdin until EOF.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunWithMgr(func(ctx context.Context, mgr *session.Manager) error {
				msg, err := resolveChatMessage(cmd, args)
				if err != nil {
					return err
				}
				modelOverride, _ := cmd.Flags().GetString("model")
				aiClient, err := localAIClient(mgr, modelOverride)
				if err != nil {
					return err
				}
				return aimode.RunAgentOnce(ctx, aimode.Options{
					AI:     aiClient,
					Scope:  pkgai.ScopeLocal,
					Out:    os.Stdout,
					ErrOut: os.Stderr,
					Local:  aimode.DefaultLocalRunner(),
				}, msg)
			})
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteAIAgentHelp)
	return cmd
}
