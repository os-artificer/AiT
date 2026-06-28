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

package validate

import (
	"github.com/os-artificer/AiT/internal/aissh/config/shared"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "validate"

// New builds the `aitssh config validate` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Validate config.yaml syntax and values",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := pkgconfig.Validate("")
			if err != nil {
				return err
			}
			return shared.EmitValidationResult(cmd, report)
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
