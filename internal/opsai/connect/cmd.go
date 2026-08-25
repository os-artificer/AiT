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

package connect

import (
	"context"

	"github.com/os-artificer/AiT/internal/opsai/deps"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

const name = "connect"

// New builds the `ai connect` cobra command.
func New(d deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "connect [ssh-options] destination",
		Short:         "Open a remote Ops Agent session over SSH",
		Long:          pkghelpdoc.LongHint(toolname.AICLI("connect")),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunWithMgr(func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error {
				return Run(ctx, cfg, aiClient, args, d.Exec)
			})
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
