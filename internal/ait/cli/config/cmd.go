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

// Package config implements the `ait config` Cobra subtree (set/get/init/view/edit/validate).
package config

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"

	"github.com/spf13/cobra"
)

// New builds the `ait config` cobra command tree.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage AiT configuration",
	}
	cmd.AddCommand(newSetCmd(), newGetCmd(), newInitCmd(), newViewCmd(), newEditCmd(), newValidateCmd())
	helpdoc.Bind(cmd, helpdoc.WriteConfigHelp)
	return cmd
}

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if provider, ok := parseAPIKeyConfigKey(key); ok {
				if err := setProviderAPIKey(provider, value); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), apiKeyStoredMessage(provider))
				return nil
			}

			cfg, err := pkgconfig.Load()
			if err != nil {
				return err
			}
			if err := cfg.SetKey(key, value); err != nil {
				return err
			}
			return pkgconfig.Save(cfg)
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteConfigSetHelp)
	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if provider, ok := parseAPIKeyConfigKey(key); ok {
				v, err := getProviderAPIKey(provider)
				if err != nil {
					return err
				}
				fmt.Println(maskSecret(v))
				return nil
			}

			cfg, err := pkgconfig.Load()
			if err != nil {
				return err
			}
			v, err := cfg.GetKey(key)
			if err != nil {
				return err
			}
			fmt.Println(v)
			return nil
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteConfigGetHelp)
	return cmd
}

func newInitCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create or update config.yaml with defaults for missing keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := pkgconfig.EnsureFile(dryRun)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case res.Created && !dryRun:
				fmt.Fprintf(out, "Created %s\n", res.Path)
			case res.Created && dryRun:
				fmt.Fprintf(out, "Would create %s\n", res.Path)
			case res.Updated:
				if dryRun {
					fmt.Fprintf(out, "Would update %s (added: %s)\n", res.Path, strings.Join(res.AddedKeys, ", "))
				} else if len(res.AddedKeys) > 0 {
					fmt.Fprintf(out, "Updated %s (added: %s)\n", res.Path, strings.Join(res.AddedKeys, ", "))
				} else {
					fmt.Fprintf(out, "Updated %s\n", res.Path)
				}
			default:
				fmt.Fprintf(out, "Config at %s is up to date\n", res.Path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print changes without writing config.yaml")
	helpdoc.Bind(cmd, helpdoc.WriteConfigInitHelp)
	return cmd
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "..."
}
