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
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/sshargs"
)

// parseShellLine splits a REPL line into words, respecting single-quoted segments.
func parseShellLine(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}
	var args []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '\'' && !inQuote:
			inQuote = true
		case c == '\'' && inQuote:
			inQuote = false
		case (c == ' ' || c == '\t') && !inQuote:
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unclosed quote")
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args, nil
}

// parseLoginArgs parses login argv via sshargs (OpenSSH-compatible).
func parseLoginArgs(args []string) (host, alias, user, identity, password string, port int, tmux bool, err error) {
	spec, err := sshargs.Parse(args)
	if err != nil {
		return "", "", "", "", "", 0, false, err
	}
	resolved, err := sshargs.Resolve(spec)
	if err != nil {
		return "", "", "", "", "", 0, false, err
	}
	return resolved.Host, resolved.Alias, resolved.User, resolved.IdentityPath,
		resolved.Password, resolved.Port, resolved.Tmux, nil
}

// parseLoadArgs parses load arguments: load <id|alias> [--tmux] [--password P].
func parseLoadArgs(args []string) (idOrAlias, password string, tmux bool, err error) {
	if len(args) < 1 {
		return "", "", false, fmt.Errorf("usage: load <id|alias> [--tmux] [--password P]")
	}
	idOrAlias = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--tmux":
			tmux = true
		case "--password":
			i++
			if i >= len(args) {
				return "", "", false, fmt.Errorf("--password requires a value")
			}
			password = args[i]
		default:
			return "", "", false, fmt.Errorf("unknown load flag %q", args[i])
		}
	}
	return idOrAlias, password, tmux, nil
}
