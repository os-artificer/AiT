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

package edit

import (
	"github.com/os-artificer/AiT/internal/aissh/config/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/platform"

	"github.com/spf13/cobra"
)

const name = "edit"

// New builds the `aitssh config edit` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Edit config.yaml in $VISUAL or $EDITOR",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := shared.EnsureConfigFileIfMissing(); err != nil {
				return err
			}

			path, err := platform.ConfigPath()
			if err != nil {
				return err
			}

			if err := shared.OpenEditorFn(path); err != nil {
				return shared.WrapEditorErr(err)
			}

			return shared.FinishConfigValidation(cmd, path)
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
