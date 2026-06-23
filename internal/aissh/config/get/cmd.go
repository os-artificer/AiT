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

package get

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/aissh/config/shared"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "get"

// New builds the `ai-ssh config get` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "get <key>",
		Short:         "Get config value",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if provider, ok := shared.ParseAPIKeyConfigKey(key); ok {
				v, err := shared.GetProviderAPIKey(provider)
				if err != nil {
					return err
				}
				fmt.Println(shared.MaskSecret(v))
				return nil
			}

			cfg, err := pkgconfig.Load()
			if err != nil {
				return err
			}
			v, err := cfg.GetKey(key)
			if err != nil {
				return err
			}
			fmt.Println(v)
			return nil
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
