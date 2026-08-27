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
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func runCommand(root, hostname, cmdline string) (stdout, stderr string, code int) {
	fields := splitCommand(strings.TrimSpace(cmdline))
	if len(fields) == 0 {
		return "", "", 0
	}

	switch fields[0] {
	case "echo":
		return strings.Join(fields[1:], " ") + "\n", "", 0
	case "true":
		return "", "", 0
	case "false":
		return "", "", 1
	case "pwd":
		return root + "\n", "", 0
	case "hostname":
		return hostname + "\n", "", 0
	case "ls":
		return runLs(root, fields[1:])
	case "cat":
		return runCat(root, fields[1:])
	default:
		return "", fields[0] + ": command not found\n", 127
	}
}

func runLs(root string, args []string) (string, string, int) {
	dir := root
	if len(args) > 0 {
		dir = resolveMockPath(root, args[0])
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err.Error() + "\n", 1
	}
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e.Name())
		b.WriteByte('\n')
	}
	return b.String(), "", 0
}

func runCat(root string, args []string) (string, string, int) {
	if len(args) == 0 {
		return "", "cat: missing operand\n", 1
	}
	data, err := os.ReadFile(resolveMockPath(root, args[0]))
	if err != nil {
		return "", "cat: " + args[0] + ": No such file or directory\n", 1
	}
	return string(data), "", 0
}

func resolveMockPath(root, p string) string {
	cleaned := filepath.Clean(filepath.FromSlash(p))
	if filepath.IsAbs(cleaned) {
		rel := strings.TrimPrefix(filepath.ToSlash(cleaned), "/")
		return filepath.Join(root, filepath.FromSlash(rel))
	}
	return filepath.Join(root, cleaned)
}

func splitCommand(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuote := false
	for _, r := range line {
		switch {
		case r == '\'' && !inQuote:
			inQuote = true
		case r == '\'' && inQuote:
			inQuote = false
		case unicode.IsSpace(r) && !inQuote:
			if cur.Len() > 0 {
				fields = append(fields, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		fields = append(fields, cur.String())
	}
	return fields
}
