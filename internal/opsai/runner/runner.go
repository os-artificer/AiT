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

package runner

import (
	"fmt"
	"io"
	"strings"

	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"

	"github.com/spf13/cobra"
)

// NewAIClient builds an AI client from cfg, optionally overriding the cloud provider.
func NewAIClient(cfg *config.Config, modelOverride string) (*pkgai.Client, error) {
	var opts []pkgai.ClientOption
	if modelOverride != "" {
		opts = append(opts, pkgai.WithProviderOverride(modelOverride))
	}
	return pkgai.NewClient(cfg, opts...)
}

// ResolveChatMessage reads the one-shot message from args or stdin.
func ResolveChatMessage(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		msg := strings.TrimSpace(strings.Join(args, " "))
		if msg == "" {
			return "", pkgai.ChatMessageRequiredErr()
		}
		return msg, nil
	}

	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		return "", pkgai.ChatMessageRequiredErr()
	}
	return msg, nil
}
