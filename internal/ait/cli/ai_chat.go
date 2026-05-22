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

package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/aimode"
	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newAIChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [message...]",
		Short: "Send one message to the AI and print the reply (non-interactive)",
		Long:  "Run 'ait ai chat --help' for usage. With no arguments, reads the prompt from stdin until EOF.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				msg, err := resolveChatMessage(cmd, args)
				if err != nil {
					return err
				}
				modelOverride, _ := cmd.Flags().GetString("model")
				aiClient, err := localAIClient(mgr, modelOverride)
				if err != nil {
					return err
				}
				return aimode.RunChatOnce(ctx, localAIModeOptions(aiClient), msg)
			})
		},
	}
	bindHelp(cmd, helpdoc.WriteAIChatHelp)
	return cmd
}

func resolveChatMessage(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		msg := strings.TrimSpace(strings.Join(args, " "))
		if msg == "" {
			return "", aimode.ErrChatMessageRequired
		}
		return msg, nil
	}

	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		return "", aimode.ErrChatMessageRequired
	}
	return msg, nil
}
