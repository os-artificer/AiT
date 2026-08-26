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

package stop

import (
	"context"
	"fmt"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgollama "github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

const name = "stop"

// New builds the `ai ollama stop` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Stop local ollama serve",
		Long:          pkghelpdoc.LongHint("ai ollama stop"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pkgollama.StopServe(context.Background()); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "stopped ollama (or was not running)")
			return nil
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
