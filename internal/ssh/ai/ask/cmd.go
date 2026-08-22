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

package ask

import (
	"context"
	"os"

	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/aimode/chat"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/ssh/ai/deps"
	"github.com/os-artificer/AiT/internal/ssh/ai/runner"

	"github.com/spf13/cobra"
)

const name = chat.Verb

// New builds the `aissh ai ask` cobra command.
func New(d deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ask [message...]",
		Short:         "Send one message to the AI and print the reply (non-interactive)",
		Long:          pkghelpdoc.LongHintChatOnce("ai ask"),
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
				return chat.RunOnce(ctx, chat.Input{
					AI:      aiClient,
					Message: msg,
					Scope:   pkgai.ScopeLocal,
					Out:     os.Stdout,
				})
			})
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
