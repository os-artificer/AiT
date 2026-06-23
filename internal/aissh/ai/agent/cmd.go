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

package agent

import (
	"context"

	"github.com/os-artificer/AiT/internal/aissh/ai/deps"
	"github.com/os-artificer/AiT/internal/aissh/ai/mode"
	"github.com/os-artificer/AiT/internal/aissh/ai/mode/chat"
	"github.com/os-artificer/AiT/internal/aissh/ai/runner"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

const name = chat.AgentVerb

// New builds the `aitssh ai agent` cobra command.
func New(d deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "agent [message...]",
		Short:         "Send one message to the AI in agent mode and execute tool calls (non-interactive)",
		Long:          pkghelpdoc.LongHintChatOnce("ai agent"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunWithMgr(func(ctx context.Context, mgr *session.Manager) error {
				msg, err := runner.ResolveChatMessage(cmd, args)
				if err != nil {
					return err
				}
				modelOverride, _ := cmd.Flags().GetString("model")
				aiClient, err := runner.LocalAIClient(mgr, modelOverride)
				if err != nil {
					return err
				}
				return mode.RunAgentOnce(ctx, runner.LocalAIModeOptions(aiClient, mgr.Config.AgentMaxSteps(), d.Exec), msg)
			})
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
