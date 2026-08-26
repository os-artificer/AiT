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

// Package model builds the `ai model` Cobra subtree.
package model

import (
	"context"
	"fmt"

	"github.com/os-artificer/AiT/internal/opsai/deps"
	"github.com/os-artificer/AiT/internal/opsai/model/subregistry"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "model"

// New builds the `ai model` cobra command.
func New(d deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Show or list configured AI providers",
		Long:          pkghelpdoc.LongHint("ai model"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunWithMgr(func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error {
				if aiClient == nil {
					return pkgai.NotConfiguredErr()
				}
				_, err := fmt.Fprintln(cmd.OutOrStdout(), aiClient.StatusLine())
				return err
			})
		},
	}
	subregistry.Attach(cmd, d)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
