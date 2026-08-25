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

// Package ollama builds the `aissh ai ollama` Cobra subtree (status/setup/start/stop/disable).
package ollama

import (
	"github.com/os-artificer/AiT/internal/opsai/ollama/subregistry"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

// New builds the `aissh ai ollama` cobra command tree.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ollama",
		Short:         "Local Ollama setup, status, and service control",
		Long:          pkghelpdoc.LongHint("ai ollama"),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	subregistry.Attach(cmd)
	pkghelpdoc.Bind(cmd, WriteAIOllamaHelp)
	return cmd
}
