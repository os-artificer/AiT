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

package ait

import (
	"context"
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/sshclient"
)

// PrintShellHelp writes help for the ait shell REPL (AiT> prompt).
func PrintShellHelp(w io.Writer) {
	const help = `AiT shell — type commands directly, press Enter to run (no !ait prefix):

  help              show this help
  exit, quit        leave AiT shell

  login --host H [--user U] [--port 22] [--alias A] [--identity PATH] [--password P] [--tmux]
  load <id|alias> [--tmux] [--password P]
  list              list saved sessions
  config, sync, session, suspend  — same as: ait <command> ...

Prompts:
  AiT>              not connected (this shell)
  AiT user@host>    remote SSH session
  AiT Local>        local mode inside SSH (!ait local)

Examples:
  login --host 192.168.1.1 --user root
  load myserver
  list
`
	shellWriteTTY(w, help)
}

// ExecuteShellLine runs one REPL line. Returns exitREPL true when the user typed exit/quit.
func ExecuteShellLine(ctx context.Context, mgr *session.Manager, line string) (exitREPL bool, err error) {
	line = normalizeShellLine(line)
	args, err := parseShellLine(line)
	if err != nil {
		return false, err
	}
	if len(args) == 0 {
		return false, nil
	}
	switch args[0] {
	case "exit", "quit":
		return true, nil
	case "help", "-h", "--help":
		PrintShellHelp(shellCLIOut)
		return false, nil
	case "login":
		return false, shellLogin(ctx, mgr, args[1:])
	case "load":
		return false, shellLoad(ctx, mgr, args[1:])
	default:
		return false, runCLI(args)
	}
}

func shellLogin(ctx context.Context, mgr *session.Manager, args []string) error {
	host, alias, user, identity, password, port, tmux, err := parseLoginArgs(args)
	if err != nil {
		return err
	}
	if host == "" {
		return fmt.Errorf("--host is required")
	}
	return mgr.Login(ctx, session.LoginOpts{
		Alias: alias, Host: host, Port: port, User: user,
		IdentityPath: identity, Password: password, Tmux: tmux,
	})
}

func shellLoad(ctx context.Context, mgr *session.Manager, args []string) error {
	id, password, tmux, err := parseLoadArgs(args)
	if err != nil {
		return err
	}
	return mgr.Load(ctx, id, tmux, password)
}

// runCLI executes cobra subcommands (shared with ExecuteMeta, without SSH-only blocks).
func runCLI(args []string) error {
	root := newRootCmd()
	root.SetOut(shellCLIOut)
	root.SetErr(shellCLIErr)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

// shellPrompt returns the REPL prompt string.
func shellPrompt() string {
	return sshclient.BasePrompt
}
