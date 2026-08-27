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

func TestKeyringDir_UnderDataDir(t *testing.T) {
	isolatePlatformHome(t)

	dir, err := KeyringDir()
	if err != nil {
		t.Fatal(err)
	}

	data, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(data, "keyring")
	if dir != want {
		t.Fatalf("KeyringDir=%q want %q", dir, want)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("keyring path is not a directory")
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("mode=%o want 0700", info.Mode().Perm())
	}
	if !strings.HasPrefix(dir, data) {
		t.Fatalf("dir %q not under DataDir %q", dir, data)
	}
}
