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
			toolname.AICLI(`agent "check disk"`),
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

func toolboxArgvLine(argv []string) string {
	return strings.Join(argv, " ")
}
