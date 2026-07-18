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

package client

import (
	"net"
	"strings"
)

const (
	nestedMOTDLastLogin     = "last login:"
	nestedMOTDWelcome       = "welcome to"
	nestedPasswordPromptSub = "password:"

	promptClearToEOL = "\x1b[K"
)

type nestedPromptState int

const (
	nestedPromptIdle nestedPromptState = iota
	nestedPromptAwaitingShell
	nestedPromptInjecting
)

// parseSSHCommandLine extracts user and host from a simple interactive ssh command line.
// Returns ok=false for non-ssh lines, ProxyJump (-J), or unparseable forms.
func parseSSHCommandLine(line, defaultUser string) (user, host string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ssh ") {
		return "", "", false
	}

	args := strings.Fields(line)
	if len(args) < 2 {
		return "", "", false
	}

	user = strings.TrimSpace(defaultUser)
	host = ""
	i := 1
	for i < len(args) {
		a := args[i]
		lower := strings.ToLower(a)

		if lower == "-j" || strings.HasPrefix(lower, "-j") || lower == "-w" || strings.HasPrefix(lower, "-w") {
			return "", "", false
		}

		if lower == "-p" || lower == "-l" || lower == "-i" || lower == "-o" || lower == "-b" || lower == "-c" {
			i += 2
			if i > len(args) {
				return "", "", false
			}
			continue
		}

		if strings.HasPrefix(lower, "-p") && len(a) > 2 {
			i++
			continue
		}

		if strings.HasPrefix(a, "-") {
			i++
			continue
		}

		dest := a
		if at := strings.LastIndex(dest, "@"); at >= 0 {
			user = dest[:at]
			host = dest[at+1:]
		} else {
			host = dest
		}
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host == "" {
			return "", "", false
		}
		if user == "" {
			user = defaultUser
		}
		return user, host, true
	}

	return "", "", false
}

func scrollChunkLooksLikePasswordPrompt(chunk []byte) bool {
	lower := strings.ToLower(string(chunk))
	return strings.Contains(lower, nestedPasswordPromptSub)
}

func scrollChunkLooksLikeMOTD(chunk []byte) bool {
	lower := strings.ToLower(string(chunk))
	return strings.Contains(lower, nestedMOTDLastLogin) || strings.Contains(lower, nestedMOTDWelcome)
}

func lineLooksLikeShellPrompt(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	lower := strings.ToLower(line)
	if strings.Contains(lower, nestedMOTDLastLogin) || strings.Contains(lower, nestedMOTDWelcome) {
		return false
	}
	if strings.Contains(lower, nestedPasswordPromptSub) {
		return false
	}
	if strings.HasPrefix(line, "AiT ") && strings.HasSuffix(strings.TrimRight(line, " \t\r"), ">") {
		return false
	}

	for _, suffix := range []string{"->", "$", "#", "%", ">", "→", "➜", "❯"} {
		if strings.HasSuffix(line, suffix) {
			return true
		}
	}

	return false
}

func scrollChunkLooksLikeShellPrompt(chunk []byte) bool {
	text := strings.ReplaceAll(string(chunk), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	for _, line := range strings.Split(text, "\n") {
		// Prefer ANSI-aware heuristics (OMZ/P10k); keep the simpler suffix match as fallback.
		if isLikelyShellPromptLine(line) || lineLooksLikeShellPrompt(line) {
			return true
		}
	}
	return false
}

// parseSSHConnectionClosedHost extracts the host from OpenSSH client text
// "Connection to HOST closed." Returns empty when the chunk is unrelated.
func parseSSHConnectionClosedHost(chunk string) string {
	chunk = strings.ReplaceAll(chunk, "\r\n", "\n")
	chunk = strings.ReplaceAll(chunk, "\r", "\n")
	for _, line := range strings.Split(chunk, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "Connection to "
		const suffix = " closed."
		if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
			continue
		}
		host := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
		return strings.TrimSpace(host)
	}
	return ""
}

// hostsEquivalent reports whether two SSH host labels refer to the same target.
func hostsEquivalent(a, b string) bool {
	return hostnameBase(a) == hostnameBase(b)
}

func hostnameBase(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if i := strings.IndexByte(host, '.'); i > 0 {
		return host[:i]
	}
	return host
}
