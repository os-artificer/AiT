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

package ollama

import (
	"context"
	"fmt"
	"runtime"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// CommandRunner runs an external command (injectable for tests).
type CommandRunner interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, name string, args []string) error
}

type osRunner struct{}

// LookPath returns the path to name using the platform's PATH lookup.
func (osRunner) LookPath(name string) (string, error) {
	return platform.LookPath(name)
}

// Run executes name with args under ctx, returning any error.
func (osRunner) Run(ctx context.Context, name string, args []string) error {
	return runCommand(ctx, name, args, "")
}

// EnsureBinary installs Ollama when allowed and missing from PATH.
func EnsureBinary(ctx context.Context, runner CommandRunner, allowed bool) error {
	if !allowed {
		return nil
	}
	if runner == nil {
		runner = osRunner{}
	}
	if _, err := runner.LookPath(binaryName); err == nil {
		return nil
	}
	return installForOS(ctx, runner)
}

func installForOS(ctx context.Context, runner CommandRunner) error {
	switch runtime.GOOS {
	case "darwin":
		if _, err := runner.LookPath("brew"); err != nil {
			return fmt.Errorf("homebrew not found; install Ollama from https://ollama.com/download")
		}
		ctx, cancel := context.WithTimeout(ctx, installTimeout)
		defer cancel()
		return runner.Run(ctx, "brew", []string{"install", "ollama"})
	case "linux":
		if _, err := runner.LookPath("curl"); err != nil {
			return fmt.Errorf("curl not found; install Ollama manually: curl -fsSL %s | sh (may need sudo)", installScriptURL)
		}
		ctx, cancel := context.WithTimeout(ctx, installTimeout)
		defer cancel()
		return runner.Run(ctx, "sh", []string{"-c", "curl -fsSL " + installScriptURL + " | sh"})
	case "windows":
		if _, err := runner.LookPath("winget"); err != nil {
			return fmt.Errorf("winget not found; download Ollama from %s", windowsDownloadURL)
		}
		ctx, cancel := context.WithTimeout(ctx, installTimeout)
		defer cancel()
		return runner.Run(ctx, "winget", []string{
			"install", "Ollama.Ollama",
			"--accept-package-agreements",
			"--accept-source-agreements",
		})
	default:
		return fmt.Errorf("unsupported OS %s; install Ollama from https://ollama.com/download", runtime.GOOS)
	}
}

// installPromptSummary returns a one-line description of the install command for this OS.
func installPromptSummary() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Ollama via Homebrew (brew install ollama)? [y/N] "
	case "linux":
		return "Install Ollama via https://ollama.com/install.sh (may need sudo)? [y/N] "
	case "windows":
		return "Install Ollama via winget (Ollama.Ollama)? [y/N] "
	default:
		return "Install Ollama from https://ollama.com/download? [y/N] "
	}
}
