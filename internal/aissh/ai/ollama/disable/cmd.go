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

package disable

import (
	"context"
	"fmt"

	provshared "github.com/os-artificer/AiT/internal/aissh/ai/provider/shared"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgollama "github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

const name = "disable"

// New builds the `ai-ssh ai ollama disable` cobra command.
func New() *cobra.Command {
	var withStop bool
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Disable Ollama provider in config",
		Long:          pkghelpdoc.LongHint("ai ollama disable"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := pkgollama.DisableProvider(context.Background(), cfg, withStop); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), provshared.MsgDisabledFmt, config.ProviderOllama)
			return nil
		},
	}
	cmd.Flags().BoolVar(&withStop, "stop", false, "Also stop ollama serve")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
