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

package config

import (
	"errors"
	"os"
	"runtime"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"

	"github.com/spf13/cobra"
)

// ErrValidationFailed indicates `ait config validate` saw at least one issue.
var ErrValidationFailed = errors.New("config validation failed")

// OpenEditorFn is the editor launcher; tests override this to skip launching $EDITOR.
var OpenEditorFn = platform.OpenEditor

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Print config.yaml contents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := ensureConfigFileIfMissing(); err != nil {
				return err
			}

			_, content, err := pkgconfig.ReadRawFile()
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(content)
			return err
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteConfigViewHelp)
	return cmd
}

func newEditCmd() *cobra.Command {
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

			if err := OpenEditorFn(path); err != nil {
				return wrapEditorErr(err)
			}

			return finishConfigValidation(cmd, path)
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteConfigEditHelp)
	return cmd
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config.yaml syntax and values",
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := pkgconfig.Validate("")
			if err != nil {
				return err
			}
			return emitValidationResult(cmd, report)
		},
	}
	helpdoc.Bind(cmd, helpdoc.WriteConfigValidateHelp)
	return cmd
}

func finishConfigValidation(cmd *cobra.Command, path string) error {
	report, err := pkgconfig.Validate(path)
	if err != nil {
		return err
	}
	return emitValidationResult(cmd, report)
}

func emitValidationResult(cmd *cobra.Command, report *pkgconfig.ValidationReport) error {
	if report.OK() {
		pkgconfig.FormatValidation(cmd.OutOrStdout(), report)
		return nil
	}

	pkgconfig.FormatValidation(cmd.ErrOrStderr(), report)
	return ErrValidationFailed
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
	_, err = pkgconfig.EnsureFileForEdit()
	return err
}

func wrapEditorErr(err error) error {
	if !errors.Is(err, platform.ErrNoEditor) {
		return err
	}
	if runtime.GOOS == "windows" {
		return errhint.WithTryMsg(
			platform.ErrNoEditor.Error(),
			"Set EDITOR or use a GUI editor.",
			"set EDITOR=notepad",
			"set EDITOR=code --wait",
		)
	}
	return errhint.WithTryMsg(
		platform.ErrNoEditor.Error(),
		"Set EDITOR or install vi.",
		"export EDITOR=vi",
		"export EDITOR=\"code --wait\"",
	)
}
