//go:build !windows

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
)

// OpenEditor launches the configured editor on path and blocks until it exits.
// path may contain spaces; it is shell-quoted for sh -c.
func OpenEditor(path string) error {
	editor, err := ResolveEditor()
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", editor+" "+shellQuote(path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultEditor() (string, error) {
	if _, err := LookPath("vi"); err == nil {
		return "vi", nil
	}
	return "", editorNotFoundErr()
}
