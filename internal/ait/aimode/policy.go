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

package aimode

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

const confirmYes = "y"

// confirmExec returns whether the user approved running cmd under opt's scope and policy.
// sessionTrusted is updated when the user accepts session_once trust.
func confirmExec(opt Options, cmd string, sessionTrusted *bool) (bool, error) {
	if opt.ExecPolicy == ExecPolicySessionOnce && *sessionTrusted {
		return true, nil
	}

	prompt := confirmPrompt(opt, cmd)
	if opt.ConfirmRead != nil {
		ok, err := opt.ConfirmRead(prompt)
		if err != nil {
			return false, err
		}
		if ok && opt.ExecPolicy == ExecPolicySessionOnce {
			*sessionTrusted = true
		}
		return ok, nil
	}

	_, _ = io.WriteString(opt.Out, prompt)
	line, err := readConfirmLine(opt.In)
	if err != nil {
		return false, err
	}

	ok := strings.EqualFold(strings.TrimSpace(line), confirmYes)
	if ok && opt.ExecPolicy == ExecPolicySessionOnce {
		*sessionTrusted = true
	}
	return ok, nil
}

func confirmPrompt(opt Options, cmd string) string {
	label := "local"
	if opt.Scope == ai.ScopeRemote {
		label = "remote"
		if opt.HostLabel != "" {
			label = opt.HostLabel
		}
	}

	if opt.ExecPolicy == ExecPolicySessionOnce {
		return fmt.Sprintf("AiT: [%s] run: %s\nTrust commands for this AI session? [y/N] ", label, cmd)
	}
	return fmt.Sprintf("AiT: [%s] run: %s\nRun? [y/N] ", label, cmd)
}

func readConfirmLine(in io.Reader) (string, error) {
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", nil
	}
	return sc.Text(), nil
}
