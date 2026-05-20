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
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandSSHPathTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	got := ExpandSSHPath("~/.ssh/id_rsa")
	want := filepath.Join(home, ".ssh", "id_rsa")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSSHConfigDir(t *testing.T) {
	dir, err := SSHConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, filepath.Join(".ssh")) {
		t.Errorf("unexpected dir %q", dir)
	}
}

func TestOpenSSHPath(t *testing.T) {
	_, err := OpenSSHPath()
	if err != nil {
		t.Skip("ssh not installed:", err)
	}
}
