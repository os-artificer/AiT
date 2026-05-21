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
	"os"
	"path/filepath"
)

func prepareZshLocalInit() (*localShellInit, []string, []string, error) {
	dir, err := os.MkdirTemp("", "ait-local-zsh-*")
	if err != nil {
		return nil, nil, nil, err
	}

	// Re-apply PROMPT after user rc so themes do not win permanently.
	content := "export PROMPT='" + LocalPrompt + "'\n" +
		"[ -f \"${HOME}/.zshrc\" ] && source \"${HOME}/.zshrc\"\n" +
		"export PROMPT='" + LocalPrompt + "'\n"

	if err := os.WriteFile(filepath.Join(dir, ".zshrc"), []byte(content), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return nil, nil, nil, err
	}
	return &localShellInit{dir: dir}, nil, []string{"ZDOTDIR=" + dir}, nil
}

func prepareBashLocalInit() (*localShellInit, []string, []string, error) {
	dir, err := os.MkdirTemp("", "ait-local-bash-*")
	if err != nil {
		return nil, nil, nil, err
	}

	rc := filepath.Join(dir, "ait-local.bashrc")
	content := "export PS1='" + localPromptPS1() + "'\n"

	if err := os.WriteFile(rc, []byte(content), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return nil, nil, nil, err
	}
	return &localShellInit{dir: dir}, []string{"--rcfile", rc}, nil, nil
}
