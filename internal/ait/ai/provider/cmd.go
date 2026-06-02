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

// Package provider builds the `ait ai provider` Cobra subtree.
package provider

import (
	_ "github.com/os-artificer/AiT/internal/ait/ai/provider/registry/all"

	"github.com/os-artificer/AiT/internal/ait/ai/deps"
	"github.com/os-artificer/AiT/internal/ait/ai/provider/subregistry"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "provider"

// New builds the `ait ai provider` cobra command.
func New(_ deps.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Enable or disable AI provider routing",
		Long:          "Run 'ait ai provider --help' for subcommands.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	subregistry.Attach(cmd)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
