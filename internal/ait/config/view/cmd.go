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

package view

import (
	"github.com/os-artificer/AiT/internal/ait/config/shared"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "view"

// New builds the `ait config view` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Print config.yaml contents",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := shared.EnsureConfigFileIfMissing(); err != nil {
				return err
			}

			_, content, err := pkgconfig.ReadRawFile()
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(content)
			return err
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
