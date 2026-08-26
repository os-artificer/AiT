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

package setup

import (
	"context"

	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgollama "github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

const (
	name              = "setup"
	flagNameBaseURL   = "base-url"
	flagNameModelsDir = "models-dir"
	flagNameNoInstall = "no-install"
	flagNameNoPull    = "no-pull"
	flagNameYes       = "yes"
	flagShorthandYes  = "y"
)

// New builds the `ai ollama setup` cobra command.
func New() *cobra.Command {
	var (
		model     string
		baseURL   string
		modelsDir string
		assumeYes bool
		noInstall bool
		noPull    bool
		dryRun    bool
	)
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Enable tiered routing and set up local Ollama",
		Long:          pkghelpdoc.LongHint("ai ollama setup"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			opts := pkgollama.Options{
				Model:     model,
				BaseURL:   baseURL,
				ModelsDir: modelsDir,
				AssumeYes: assumeYes,
				NoInstall: noInstall,
				NoPull:    noPull,
				DryRun:    dryRun,
				Out:       cmd.OutOrStdout(),
				Err:       cmd.ErrOrStderr(),
				In:        cmd.InOrStdin(),
			}
			_, err = pkgollama.RunSetup(context.Background(), cfg, opts)
			if err != nil {
				return err
			}
			if !dryRun {
				return config.Save(cfg)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&model, "model", config.DefaultOllamaModel, "Ollama model to pull")
	cmd.Flags().StringVar(&baseURL, flagNameBaseURL, config.DefaultOllamaBase, "Ollama OpenAI-compatible base URL")
	cmd.Flags().StringVar(&modelsDir, flagNameModelsDir, "", "Directory for Ollama model blobs (OLLAMA_MODELS); prompts when unset")
	cmd.Flags().BoolVarP(&assumeYes, flagNameYes, flagShorthandYes, false, "Allow Ollama install without prompting")
	cmd.Flags().BoolVar(&noInstall, flagNameNoInstall, false, "Never install Ollama (same as declining the prompt)")
	cmd.Flags().BoolVar(&noPull, flagNameNoPull, false, "Skip ollama pull")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print planned steps only")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
