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

package opsai

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/os-artificer/AiT/internal/opsai/deps"
	_ "github.com/os-artificer/AiT/internal/opsai/registry/all"
	"github.com/os-artificer/AiT/internal/opsai/version"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Execute runs the top-level ai CLI with args on os.Stdout and os.Stderr.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level ai CLI with custom stdout and stderr.
func ExecuteWriters(args []string, out, errOut io.Writer) error {
	if len(args) == 1 && args[0] == aitplugin.ManifestArg {
		return aitplugin.WriteManifest(out, toolname.AI)
	}

	args, saw := config.StripAIExperimentalFlag(args)
	if saw {
		config.SetAIExperimentalCLI(true)
	}

	if pkghelpdoc.HelpRequested(args) {
		return WriteAIHelp(out)
	}
	if len(args) == 1 && (args[0] == cmdVersion || args[0] == "-V") {
		_, err := fmt.Fprintln(out, version.Version())
		return err
	}

	return executeCobraRoot(args, out, errOut)
}

func newRootCmd() *cobra.Command {
	d := deps.Deps{
		Exec:           execAisshCLI,
		RunWithMgr:     runWithConfig,
		ApplyOpsDomain: withOpsDomain,
	}
	return New(d)
}

func executeCobraRoot(args []string, out, errOut io.Writer) error {
	root := newRootCmd()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

func runWithConfig(fn func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error) error {
	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := CommandContext()
	defer cancel()

	aiClient, err := pkgai.NewClient(cfg)
	if err != nil {
		aiClient = nil
	}
	return fn(ctx, cfg, aiClient)
}

// execAisshCLI runs aissh from PATH with the given argv (aish_cli tool path).
func execAisshCLI(args []string, out, errOut io.Writer) error {
	bin, err := exec.LookPath(toolname.AISsh)
	if err != nil {
		return fmt.Errorf("%s not found on PATH: %w", toolname.AISsh, err)
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
