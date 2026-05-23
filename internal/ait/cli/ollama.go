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

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

func newOllamaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ollama",
		Short: "Local Ollama setup, status, and service control",
		Long:  "Run 'ait ai ollama --help' for subcommands.",
	}
	cmd.AddCommand(
		newOllamaStatusCmd(),
		newOllamaSetupCmd(),
		newOllamaStartCmd(),
		newOllamaStopCmd(),
		newOllamaDisableCmd(),
	)
	bindHelp(cmd, helpdoc.WriteAIOllamaHelp)
	return cmd
}

func newOllamaStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Ollama config and service status",
		Long:  helpdoc.LongHint("ai ollama status"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return ollama.PrintStatus(context.Background(), cfg, cmd.OutOrStdout())
		},
	}
	bindHelp(cmd, helpdoc.WriteOllamaStatusHelp)
	return cmd
}

func newOllamaSetupCmd() *cobra.Command {
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
		Use:   "setup",
		Short: "Enable tiered routing and set up local Ollama",
		Long:  helpdoc.LongHint("ai ollama setup"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			opts := ollama.Options{
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
			_, err = ollama.RunSetup(context.Background(), cfg, opts)
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
	bindHelp(cmd, helpdoc.WriteOllamaSetupHelp)
	return cmd
}

func newOllamaStartCmd() *cobra.Command {
	var wait bool
	var baseURL string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ollama serve in the background",
		Long:  helpdoc.LongHint("ai ollama start"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			url := baseURL
			if url == "" {
				url = ollama.BaseURLFromConfig(cfg)
			}
			err = ollama.StartServe(context.Background(), nil, url, wait, ollama.ModelsDirFromConfig(cfg))
			if errors.Is(err, ollama.ErrAlreadyRunning) {
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
	bindHelp(cmd, helpdoc.WriteOllamaStartHelp)
	return cmd
}

func newOllamaStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop local ollama serve",
		Long:  helpdoc.LongHint("ai ollama stop"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ollama.StopServe(context.Background()); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "stopped ollama (or was not running)")
			return nil
		},
	}
	bindHelp(cmd, helpdoc.WriteOllamaStopHelp)
	return cmd
}

func newOllamaDisableCmd() *cobra.Command {
	var withStop bool
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable Ollama provider in config",
		Long:  helpdoc.LongHint("ai ollama disable"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := ollama.DisableProvider(context.Background(), cfg, withStop); err != nil {
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
	bindHelp(cmd, helpdoc.WriteOllamaDisableHelp)
	return cmd
}

func ollamaSetupFlagsFromCmd(cmd *cobra.Command) ollama.Options {
	model, _ := cmd.Flags().GetString("model")
	baseURL, _ := cmd.Flags().GetString(flagNameBaseURL)
	modelsDir, _ := cmd.Flags().GetString(flagNameModelsDir)
	assumeYes, _ := cmd.Flags().GetBool(flagNameYes)
	noInstall, _ := cmd.Flags().GetBool(flagNameNoInstall)
	noPull, _ := cmd.Flags().GetBool(flagNameNoPull)
	return ollama.Options{
		Model:     model,
		BaseURL:   baseURL,
		ModelsDir: modelsDir,
		AssumeYes: assumeYes,
		NoInstall: noInstall,
		NoPull:    noPull,
		Out:       cmd.OutOrStdout(),
		Err:       cmd.ErrOrStderr(),
		In:        cmd.InOrStdin(),
	}
}
