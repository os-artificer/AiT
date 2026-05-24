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
	"errors"
	"os"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/platform"

	"github.com/spf13/cobra"
)

var errConfigValidationFailed = errors.New("config validation failed")

var openEditorFn = platform.OpenEditor

func newConfigViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Print config.yaml contents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := ensureConfigFileIfMissing(); err != nil {
				return err
			}

			_, content, err := config.ReadRawFile()
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(content)
			return err
		},
	}
	bindHelp(cmd, helpdoc.WriteConfigViewHelp)
	return cmd
}

func newConfigEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit config.yaml in $VISUAL or $EDITOR",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := ensureConfigFileIfMissing(); err != nil {
				return err
			}

			path, err := platform.ConfigPath()
			if err != nil {
				return err
			}

			if err := openEditorFn(path); err != nil {
				return wrapEditorErr(err)
			}

			return finishConfigValidation(cmd, path)
		},
	}
	bindHelp(cmd, helpdoc.WriteConfigEditHelp)
	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config.yaml syntax and values",
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := config.Validate("")
			if err != nil {
				return err
			}
			return emitValidationResult(cmd, report)
		},
	}
	bindHelp(cmd, helpdoc.WriteConfigValidateHelp)
	return cmd
}

func finishConfigValidation(cmd *cobra.Command, path string) error {
	report, err := config.Validate(path)
	if err != nil {
		return err
	}
	return emitValidationResult(cmd, report)
}

func emitValidationResult(cmd *cobra.Command, report *config.ValidationReport) error {
	if report.OK() {
		config.FormatValidation(cmd.OutOrStdout(), report)
		return nil
	}

	config.FormatValidation(cmd.ErrOrStderr(), report)
	return errConfigValidationFailed
}

func ensureConfigFileIfMissing() error {
	path, err := platform.ConfigPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	_, err = config.EnsureFileForEdit()
	return err
}
