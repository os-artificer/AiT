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
	"runtime"
	"strings"
	"testing"
)

// TestLoginInteractivePathDoesNotUseTransferPackage ensures aissh login/load setup
// does not pull in internal/pkg/transfer (get/put SFTP is separate from RunInteractive init).
func TestLoginInteractivePathDoesNotUseTransferPackage(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)

	files := []string{"client.go", "aliases.go", "client_init_helpers.go", filepath.Join("prompt", "prompt.go")}
	for _, name := range files {
		path := filepath.Join(dir, name)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), `"github.com/os-artificer/AiT/internal/pkg/transfer"`) {
			t.Errorf("%s must not import transfer package (login init path)", name)
		}
	}
}
