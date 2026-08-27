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

package testssh

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

var silentFirstTokens = map[string]struct{}{
	"export":  {},
	"unset":   {},
	"stty":    {},
	"typeset": {},
	"zle":     {},
	"print":   {},
	"printf":  {},
	"[":       {},
	"[[":      {},
	":":       {},
	"true":    {},
	"hash":    {},
	"bind":    {},
	"set":     {},
	"shopt":   {},
	"unalias": {},
	"cd":      {},
}

func drainSessionRequests(in <-chan *ssh.Request) {
	for req := range in {
		if !req.WantReply {
			continue
		}
		ok := req.Type == "pty-req" || req.Type == "env" || req.Type == "window-change" || req.Type == "shell"
		_ = req.Reply(ok, nil)
	}
}

func runInteractiveShell(ch ssh.Channel, cfg Config) {
	_, _ = io.WriteString(ch, mockPrompt(cfg))

	r := bufio.NewReader(ch)
	for {
		line, err := readShellLine(r)
		if handleShellLine(ch, cfg, line) {
			return
		}
		if err != nil {
			return
		}
	}
}

func handleShellLine(ch ssh.Channel, cfg Config, line string) (done bool) {
	if isExitLine(line) {
		return true
	}
	if isSilentShellLine(line) {
		// First prompt after PS1 setup. aissh may also paint one; pytest drains duplicates.
		_, _ = io.WriteString(ch, mockPrompt(cfg))
		return false
	}

	_, _ = io.WriteString(ch, line+"\r\n")
	stdout, stderr, _ := runCommand(cfg.RootDir, cfg.Hostname, strings.TrimSpace(line))
	if stdout != "" {
		_, _ = io.WriteString(ch, stdout)
	}
	if stderr != "" {
		_, _ = io.WriteString(ch.Stderr(), stderr)
	}
	_, _ = io.WriteString(ch, mockPrompt(cfg))
	return false
}

func readShellLine(r *bufio.Reader) (string, error) {
	var b strings.Builder
	for {
		c, err := r.ReadByte()
		if err != nil {
			return b.String(), err
		}
		if c == '\n' {
			return b.String(), nil
		}
		if c == '\r' {
			if next, peekErr := r.Peek(1); peekErr == nil && next[0] == '\n' {
				_, _ = r.ReadByte()
			}
			return b.String(), nil
		}
		b.WriteByte(c)
	}
}

func mockPrompt(cfg Config) string {
	dir := filepath.Base(cfg.RootDir)
	if dir == "" || dir == "." || dir == string(filepath.Separator) {
		dir = "/"
	}
	host := cfg.promptListenHost
	if host == "" {
		host = cfg.Hostname
	}
	return fmt.Sprintf("AiT %s@%s:%s> ", cfg.User, host, dir)
}

func isExitLine(line string) bool {
	tok := firstShellToken(line)
	return tok == "exit" || tok == "logout"
}

func isSilentShellLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	if strings.Contains(trimmed, "PS1=") || strings.Contains(trimmed, "PROMPT=") {
		return true
	}
	if strings.Contains(trimmed, promptReadySentinel) {
		return true
	}
	_, ok := silentFirstTokens[firstShellToken(trimmed)]
	return ok
}

func firstShellToken(line string) string {
	fields := splitCommand(strings.TrimSpace(line))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
