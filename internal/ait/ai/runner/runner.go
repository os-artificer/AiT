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

// Package runner holds shared one-shot AI command helpers (no subcommand imports).
package runner

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/ai/mode"
	"github.com/os-artificer/AiT/internal/ait/ai/mode/agent"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

// LocalAIClient returns the session AI client, optionally overriding the cloud provider.
func LocalAIClient(mgr *session.Manager, modelOverride string) (*pkgai.Client, error) {
	if modelOverride == "" && mgr.AI != nil {
		return mgr.AI, nil
	}
	var opts []pkgai.ClientOption
	if modelOverride != "" {
		opts = append(opts, pkgai.WithProviderOverride(modelOverride))
	}
	return pkgai.NewClient(mgr.Config, opts...)
}

// ResolveChatMessage reads the one-shot message from args or stdin.
func ResolveChatMessage(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		msg := strings.TrimSpace(strings.Join(args, " "))
		if msg == "" {
			return "", pkgai.ChatMessageRequiredErr()
		}
		return msg, nil
	}

	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		return "", pkgai.ChatMessageRequiredErr()
	}
	return msg, nil
}

// LocalAIModeOptions builds REPL options for local AI mode.
func LocalAIModeOptions(aiClient *pkgai.Client, maxSteps int, exec func([]string, io.Writer, io.Writer) error) mode.Options {
	return mode.Options{
		Scope:    pkgai.ScopeLocal,
		AI:       aiClient,
		Local:    mode.DefaultLocalRunner(),
		Capture:  agent.DefaultLocalCaptureRunner(),
		CLI:      func(argv []string, out, errOut io.Writer) error { return exec(argv, out, errOut) },
		In:       os.Stdin,
		Out:      os.Stdout,
		ErrOut:   os.Stderr,
		MaxSteps: maxSteps,
	}
}
