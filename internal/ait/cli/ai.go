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
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/aimode"
	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"

	"github.com/spf13/cobra"
)

type ptyRemoteRunner struct {
	stdin io.Writer
}

func (r *ptyRemoteRunner) Run(cmd string) error {
	if r.stdin == nil {
		return fmt.Errorf("remote stdin not available")
	}
	_, err := io.WriteString(r.stdin, cmd+"\n")
	return err
}

// runRemoteAIMode is wired into SSH InteractiveOpts.OnRemoteAI.
func runRemoteAIMode(p sshclient.RemoteAIParams) error {
	if p.AI == nil {
		return aimode.ErrAINotConfigured
	}

	return aimode.RunREPL(p.Ctx, aimode.Options{
		Scope:      ai.ScopeRemote,
		AI:         p.AI,
		HostLabel:  p.HostLabel,
		RemoteCwd:  p.RemoteCwd,
		Scrollback: p.Scrollback,
		Remote:     &ptyRemoteRunner{stdin: p.Stdin},
		In:         os.Stdin,
		Out:        os.Stdout,
		ErrOut:     os.Stderr,
	}, p.InitialQuestion)
}

func runLocalAIModeFromShell(ctx context.Context, mgr *session.Manager) error {
	return runLocalAIMode(ctx, mgr, "", "")
}

func newAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Interactive local AI mode (AiT AI> prompt)",
		Long:  "Run AI-assisted REPL on this machine. Inside SSH use !ait ai or !! instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			modelFlag, _ := cmd.Flags().GetString("model")
			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				return runLocalAIMode(ctx, mgr, modelFlag, strings.Join(args, " "))
			})
		},
	}
	cmd.PersistentFlags().String("model", "", "Override AI provider for this session (deepseek|minimax)")
	cmd.AddCommand(newAIChatCmd(), newOllamaCmd(), newProviderCmd())
	bindHelp(cmd, helpdoc.WriteAIHelp)
	return cmd
}

func runLocalAIMode(ctx context.Context, mgr *session.Manager, modelOverride, initialLine string) error {
	modelOverride = strings.ToLower(strings.TrimSpace(modelOverride))
	aiClient, err := localAIClient(mgr, modelOverride)
	if err != nil {
		return err
	}
	return aimode.RunREPL(ctx, localAIModeOptions(aiClient), initialLine)
}

func localAIClient(mgr *session.Manager, modelOverride string) (*ai.Client, error) {
	if modelOverride == "" && mgr.AI != nil {
		return mgr.AI, nil
	}
	var opts []ai.ClientOption
	if modelOverride != "" {
		opts = append(opts, ai.WithProviderOverride(modelOverride))
	}
	return ai.NewClient(mgr.Config, opts...)
}

func localAIModeOptions(aiClient *ai.Client) aimode.Options {
	return aimode.Options{
		Scope:  ai.ScopeLocal,
		AI:     aiClient,
		Local:  aimode.DefaultLocalRunner(),
		CLI:    func(argv []string, out, errOut io.Writer) error { return ExecuteWriters(argv, out, errOut) },
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}
