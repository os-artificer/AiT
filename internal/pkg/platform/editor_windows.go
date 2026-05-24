//go:build windows

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

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenEditor launches the configured editor on path and blocks until it exits.
func OpenEditor(path string) error {
	editor, err := ResolveEditor()
	if err != nil {
		return err
	}

	if isNotepadEditor(editor) {
		exe, err := notepadPath()
		if err != nil {
			return err
		}
		cmd := exec.Command(exe, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("cmd", "/c", editor+" "+windowsQuotePath(path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultWindowsEditor() (string, error) {
	if _, err := notepadPath(); err != nil {
		return "", editorNotFoundErr()
	}
	return "notepad", nil
}

func defaultEditor() (string, error) {
	return defaultWindowsEditor()
}

func notepadPath() (string, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	p := filepath.Join(sysRoot, "System32", "notepad.exe")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	return LookPath("notepad")
}

func isNotepadEditor(editor string) bool {
	base := strings.ToLower(filepath.Base(strings.Fields(editor)[0]))
	return base == "notepad" || base == "notepad.exe"
}

func windowsQuotePath(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
