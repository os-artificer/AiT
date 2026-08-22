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

package shell

import (
	"context"
	"errors"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/shellext"
	"github.com/os-artificer/AiT/internal/ssh/shell/builtin"
)

// ErrCommandNotFound is returned when no builtin, bin, aissh CLI, or PATH match exists.
var ErrCommandNotFound = errors.New("shell: command not found")

var externalRunner func(context.Context, string, []string) error

// SetExternalRunner registers the function that runs a resolved bin or PATH executable.
// Called from repl setup; tests may inject a mock.
func SetExternalRunner(fn func(context.Context, string, []string) error) {
	externalRunner = fn
}

func dispatchShellCommand(ctx context.Context, mgr *session.Manager, words []string) error {
	if path, ok := shellext.ResolveAppBin(words[0]); ok {
		return runExternal(ctx, path, words)
	}

	if err := runCLI(words); err == nil {
		return nil
	} else if !pkghelpdoc.IsUnknownCommand(err) {
		return err
	} else if cliroute.IsRemovedTopLevelWord(words[0]) {
		return err
	}

	sysPath, lpErr := platform.LookPath(words[0])
	if lpErr == nil {
		return runExternal(ctx, sysPath, words)
	}

	var lookup cliroute.AliasLookup
	if mgr != nil {
		lookup = mgr.Store
	}
	if cliroute.ShouldREPLSSHConnect(words, lookup) {
		if mgr != nil && mgr.HasActiveSession() {
			return builtin.LoginBlockedWhenActiveErr()
		}
		return runSSHConnect(ctx, mgr, words)
	}

	return commandNotFoundErr(words[0])
}

func runExternal(ctx context.Context, path string, argv []string) error {
	if externalRunner == nil {
		return fmt.Errorf("%s", errShellExternalRunnerNotConfig)
	}
	return externalRunner(ctx, path, argv)
}
