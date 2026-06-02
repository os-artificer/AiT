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

// Package ollama builds the `ait ai ollama` Cobra subtree (status/setup/start/stop/disable).
package ollama

import (
	"context"
	"errors"
	"fmt"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkgollama "github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

const (
	flagNameBaseURL   = "base-url"
	flagNameModelsDir = "models-dir"
	flagNameNoInstall = "no-install"
	flagNameNoPull    = "no-pull"
	flagNameYes       = "yes"
	flagShorthandYes  = "y"
)

// New builds the `ait ai ollama` cobra command tree.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ollama",
		Short:         "Local Ollama setup, status, and service control",
		Long:          "Run 'ait ai ollama --help' for subcommands.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		newStatusCmd(),
		newSetupCmd(),
		newStartCmd(),
		newStopCmd(),
		newDisableCmd(),
	)
	helpdoc.Bind(cmd, helpdoc.WriteAIOllamaHelp)
	return cmd
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show Ollama config and service status",
		Long:          helpdoc.LongHint("ai ollama status"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return pkgollama.PrintStatus(context.Background(), cfg, cmd.OutOrStdout())
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteOllamaStatusHelp)
	return cmd
}

func newSetupCmd() *cobra.Command {
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
		Use:           "setup",
		Short:         "Enable tiered routing and set up local Ollama",
		Long:          helpdoc.LongHint("ai ollama setup"),
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
	helpdoc.Bind(cmd, helpdoc.WriteOllamaSetupHelp)
	return cmd
}

func newStartCmd() *cobra.Command {
	var wait bool
	var baseURL string
	cmd := &cobra.Command{
		Use:           "start",
		Short:         "Start ollama serve in the background",
		Long:          helpdoc.LongHint("ai ollama start"),
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
	helpdoc.Bind(cmd, helpdoc.WriteOllamaStartHelp)
	return cmd
}

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "stop",
		Short:         "Stop local ollama serve",
		Long:          helpdoc.LongHint("ai ollama stop"),
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
	helpdoc.Bind(cmd, helpdoc.WriteOllamaStopHelp)
	return cmd
}

func newDisableCmd() *cobra.Command {
	var withStop bool
	cmd := &cobra.Command{
		Use:           "disable",
		Short:         "Disable Ollama provider in config",
		Long:          helpdoc.LongHint("ai ollama disable"),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := pkgollama.DisableProvider(context.Background(), cfg, withStop); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "disabled provider ollama")
			return nil
		},
	}
	cmd.Flags().BoolVar(&withStop, "stop", false, "Also stop ollama serve")
	helpdoc.Bind(cmd, helpdoc.WriteOllamaDisableHelp)
	return cmd
}
