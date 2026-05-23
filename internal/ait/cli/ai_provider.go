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
	"fmt"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/ollama"

	"github.com/spf13/cobra"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Enable or disable AI provider routing",
		Long:  "Run 'ait ai provider --help' for subcommands.",
	}
	cmd.AddCommand(newProviderEnableCmd(), newProviderDisableCmd())
	bindHelp(cmd, helpdoc.WriteAIProviderHelp)
	return cmd
}

func newProviderEnableCmd() *cobra.Command {
	var assumeYes, noInstall, noPull bool
	cmd := &cobra.Command{
		Use:   "enable <provider>",
		Short: "Enable an AI provider for routing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !config.IsKnownProvider(id) {
				return fmt.Errorf(errUnknownProviderFmt, id)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.SetProviderEnabled(id, true); err != nil {
				return err
			}

			if id == config.ProviderOllama {
				if err := ollama.ApplyTieredConfig(cfg, "", ""); err != nil {
					return err
				}
				opts := ollamaSetupFlagsFromCmd(cmd)
				opts.AssumeYes = assumeYes
				opts.NoInstall = noInstall
				opts.NoPull = noPull
				_, err = ollama.EnsureLocal(context.Background(), cfg, opts)
				if err != nil {
					_ = config.Save(cfg)
					return err
				}
			} else if _, err := creds.GetProviderAPIKey(id); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "AiT: warning: no API key for %s (%s)\n", id, apiKeyHint(id))
			}

			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), msgEnabledProviderFmt, id)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&assumeYes, flagNameYes, flagShorthandYes, false, "Allow Ollama install without prompting (ollama only)")
	cmd.Flags().BoolVar(&noInstall, flagNameNoInstall, false, "Skip Ollama install (ollama only)")
	cmd.Flags().BoolVar(&noPull, flagNameNoPull, false, "Skip ollama pull (ollama only)")
	cmd.Flags().String("model", config.DefaultOllamaModel, "Ollama model (ollama only)")
	cmd.Flags().String(flagNameBaseURL, config.DefaultOllamaBase, "Ollama API base (ollama only)")
	cmd.Flags().String(flagNameModelsDir, "", "Ollama models directory (ollama only)")
	bindHelp(cmd, helpdoc.WriteProviderEnableHelp)
	return cmd
}

func newProviderDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <provider>",
		Short: "Disable an AI provider (no requests will be sent to it)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !config.IsKnownProvider(id) {
				return fmt.Errorf(errUnknownProviderFmt, id)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.SetProviderEnabled(id, false); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), msgDisabledProviderFmt, id)
			return nil
		},
	}
	bindHelp(cmd, helpdoc.WriteProviderDisableHelp)
	return cmd
}

func apiKeyHint(provider string) string {
	return fmt.Sprintf("ait config set ai.%s.api_key <key>", provider)
}
