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

// Package agentops provides ops ContextBuilder and PATH toolbox execution for agentcore.
package agentops

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var toolboxBins = map[string]struct{}{
	toolname.AIHost: {},
	toolname.AILog:  {},
	toolname.AISsh:  {},
	toolname.AIImg:  {},
}

// DispatchToolbox runs argv[0] from PATH when it is an allowed toolbox binary.
func DispatchToolbox(argv []string, out, errOut io.Writer) error {
	if len(argv) == 0 {
		return errhint.WithTryMsg("toolbox requires a command",
			"Pass a full argv starting with aihost, ailog, aissh, or aiimg.",
			toolname.AIT+" agent \"check disk\"",
		)
	}
	name := argv[0]
	if _, ok := toolboxBins[name]; !ok {
		return errhint.WithTryMsg("unknown toolbox binary: "+name,
			"Allowed binaries: aihost, ailog, aissh, aiimg.",
			"aihost snapshot --json",
			"ailog search error",
		)
	}
	bin, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found on PATH: %w", name, err)
	}
	cmd := exec.Command(bin, argv[1:]...)
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// BuildContext collects read-only ops summaries for the first agent turn.
// Failures emit warnings via warn and return partial context.
func BuildContext(ctx context.Context, warn func(string)) string {
	var parts []string

	if s := snapshotHost(ctx); s != "" {
		parts = append(parts, s)
	} else if warn != nil {
		warn("aihost snapshot unavailable")
	}
	if s := listSessions(ctx); s != "" {
		parts = append(parts, s)
	} else if warn != nil {
		warn("aissh session list unavailable")
	}
	return strings.Join(parts, "\n")
}

func snapshotHost(ctx context.Context) string {
	bin, err := exec.LookPath(toolname.AIHost)
	if err != nil {
		return ""
	}
	cmd := exec.CommandContext(ctx, bin, "snapshot", "--json")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if len(s) > 2048 {
		s = s[:2048] + "..."
	}
	return "host snapshot:\n" + s
}

func listSessions(ctx context.Context) string {
	bin, err := exec.LookPath(toolname.AISsh)
	if err != nil {
		return ""
	}
	cmd := exec.CommandContext(ctx, bin, "session", "--list")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return ""
	}
	s := strings.TrimSpace(buf.String())
	if s == "" {
		return ""
	}
	if len(s) > 1024 {
		s = s[:1024] + "..."
	}
	return "ssh sessions:\n" + s
}
