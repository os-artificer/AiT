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
	"strings"

	"github.com/os-artificer/AiT/internal/ait/version"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// lineHandler runs one reserved REPL verb. exitREPL is true for exit/quit.
type lineHandler func(ctx context.Context, words []string, hist *terminal.LineHistory, out, errOut io.Writer) (exitREPL bool, err error)

// reservedHelpEntry is one reserved command row for root help (order matters).
type reservedHelpEntry struct {
	name string
	desc string
}

// reservedHandlers is the single registry for reserved REPL verbs (including manage + agent).
var reservedHandlers = map[string]lineHandler{}

func init() {
	registerReserved(cmdExit, "", handleExit)
	registerReserved(cmdQuit, "", handleExit)
	registerReserved(cmdHelp, "show this help or help <plugin>", handleHelp)
	registerReserved(aitplugin.CmdAgent, "ops agent task loop (experimental)", handleAgent)
	registerReserved(cmdVersion, "print ait version", handleVersion)
	registerReserved(cmdHistory, "list or clear input history", handleHistory)
	registerReserved(cmdClear, "clear terminal screen", handleClear)
	registerReserved(cmdInstall, "", handleManage(aitrelease.ActionInstall))
	registerReserved(cmdUpdate, "", handleManage(aitrelease.ActionUpdate))
	registerReserved(cmdUpgrade, "", handleManage(aitrelease.ActionUpgrade))
}

func registerReserved(name, _ string, fn lineHandler) {
	reservedHandlers[name] = fn
}

// reservedHelpEntries returns ordered reserved command rows for root help (excluding quit alias).
func reservedHelpEntries() []reservedHelpEntry {
	return []reservedHelpEntry{
		{cmdHelp, "show this help or help <plugin>"},
		{aitplugin.CmdAgent, "ops agent task loop (experimental)"},
		{cmdVersion, "print ait version"},
		{cmdHistory, "list or clear input history"},
		{cmdClear, "clear terminal screen"},
	}
}

func handleExit(context.Context, []string, *terminal.LineHistory, io.Writer, io.Writer) (bool, error) {
	return true, nil
}

func handleHelp(_ context.Context, words []string, _ *terminal.LineHistory, out, _ io.Writer) (bool, error) {
	if len(words) == 1 {
		return false, WriteRootHelp(out)
	}
	return false, writePluginHelp(out, words[1])
}

func handleVersion(_ context.Context, _ []string, _ *terminal.LineHistory, out, _ io.Writer) (bool, error) {
	_, err := fmt.Fprintln(out, version.Version())
	return false, err
}

func handleClear(_ context.Context, _ []string, _ *terminal.LineHistory, out, _ io.Writer) (bool, error) {
	terminal.ClearScreen(out)
	return false, nil
}

func handleHistory(_ context.Context, words []string, hist *terminal.LineHistory, out, _ io.Writer) (bool, error) {
	return false, terminal.RunHistoryCommand(out, hist, words[1:])
}

func handleAgent(ctx context.Context, words []string, _ *terminal.LineHistory, out, errOut io.Writer) (bool, error) {
	goal := strings.Join(words[1:], " ")
	return false, runNestedAgent(ctx, goal, out, errOut)
}

func handleManage(action aitrelease.Action) lineHandler {
	return func(ctx context.Context, words []string, _ *terminal.LineHistory, out, _ io.Writer) (bool, error) {
		return false, runManage(ctx, action, words[1:], out)
	}
}
