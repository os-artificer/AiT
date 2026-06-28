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

package cliroute

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

var hostConfiguredHook func(name string) bool

// SetHostConfiguredHook replaces HostConfigured for tests; pass nil to restore default.
func SetHostConfiguredHook(fn func(name string) bool) {
	hostConfiguredHook = fn
}

// HostConfigured reports whether name matches a Host stanza in the user's ssh config.
// The catch-all pattern Host * is ignored so REPL unknown words are not treated as connect targets.
func HostConfigured(name string) bool {
	if hostConfiguredHook != nil {
		return hostConfiguredHook(name)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	return hostConfiguredInFile(name, defaultSSHConfigPath())
}

func defaultSSHConfigPath() string {
	dir, err := platform.SSHConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "config")
}

func hostConfiguredInFile(name, configPath string) bool {
	patterns, err := sshHostPatternsInFile(configPath)
	if err != nil || len(patterns) == 0 {
		return false
	}
	for _, pat := range patterns {
		if isIgnoredHostPattern(pat) {
			continue
		}
		if hostPatternMatch(pat, name) {
			return true
		}
	}
	return false
}

func sshHostPatternsInFile(configPath string) ([]string, error) {
	if configPath == "" {
		return nil, os.ErrNotExist
	}
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "host ") {
			patterns = append(patterns, parseHostPatterns(line[5:])...)
			continue
		}
		if strings.HasPrefix(lower, "match ") {
			continue
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

func isIgnoredHostPattern(pattern string) bool {
	return strings.TrimSpace(pattern) == "*"
}

func parseHostPatterns(rest string) []string {
	fields := strings.Fields(rest)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" || strings.HasPrefix(f, "!") {
			continue
		}
		out = append(out, f)
	}
	return out
}

func hostPatternMatch(pattern, name string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	name = strings.ToLower(strings.TrimSpace(name))
	if pattern == "" || name == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	return sshGlobMatch(pattern, name)
}

func sshGlobMatch(pattern, name string) bool {
	star := strings.Index(pattern, "*")
	if star < 0 {
		return matchRunes([]rune(pattern), []rune(name))
	}

	prefix := pattern[:star]
	pattern = pattern[star+1:]
	if prefix == "" {
		if pattern == "" {
			return true
		}
		for i := 0; i <= len(name); i++ {
			if sshGlobMatch(pattern, name[i:]) {
				return true
			}
		}
		return false
	}

	idx := strings.Index(name, prefix)
	for idx >= 0 {
		if sshGlobMatch(pattern, name[idx+len(prefix):]) {
			return true
		}
		name = name[idx+1:]
		idx = strings.Index(name, prefix)
	}
	return false
}

func matchRunes(pat, name []rune) bool {
	for len(pat) > 0 {
		if len(name) == 0 {
			return false
		}
		switch pat[0] {
		case '?':
			pat = pat[1:]
			name = name[1:]
		default:
			if pat[0] != name[0] {
				return false
			}
			pat = pat[1:]
			name = name[1:]
		}
	}
	return len(name) == 0
}
