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
	"strings"

	"github.com/os-artificer/AiT/internal/opsai/deps"
	"github.com/os-artificer/AiT/internal/opsai/runner"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/aichat"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = aichat.Verb

// New builds the `ai ask` cobra command.
func New(d deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ask [message...]",
		Short:         "Send one message to the AI and print the reply (non-interactive)",
		Long:          pkghelpdoc.LongHintChatOnce("ai ask"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunWithMgr(func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error {
				msg, err := runner.ResolveChatMessage(cmd, args)
				if err != nil {
					return err
				}
				modelOverride, _ := cmd.Flags().GetString("model")
				modelOverride = strings.ToLower(strings.TrimSpace(modelOverride))
				client := aiClient
				if modelOverride != "" || client == nil {
					client, err = runner.NewAIClient(cfg, modelOverride)
					if err != nil {
						return err
					}
				}
				return aichat.RunOnce(ctx, aichat.Input{
					AI:      client,
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
