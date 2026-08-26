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

package start

import (
	"context"
	"errors"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgollama "github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

const (
	name            = "start"
	flagNameBaseURL = "base-url"
)

// New builds the `ai ollama start` cobra command.
func New() *cobra.Command {
	var wait bool
	var baseURL string
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Start ollama serve in the background",
		Long:          pkghelpdoc.LongHint("ai ollama start"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			url := baseURL
			if url == "" {
				url = pkgollama.BaseURLFromConfig(cfg)
			}
			err = pkgollama.StartServe(context.Background(), nil, url, wait, pkgollama.ModelsDirFromConfig(cfg))
			if errors.Is(err, pkgollama.ErrAlreadyRunning) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ollama API already reachable")
				return nil
			}
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "started ollama serve")
			return nil
		},
	}
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until the API responds")
	cmd.Flags().StringVar(&baseURL, flagNameBaseURL, "", "API base URL to probe (default from config)")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
