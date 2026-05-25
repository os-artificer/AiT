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

package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLineHistory_LoadMissingOK(t *testing.T) {
	h := NewLineHistory()
	path := filepath.Join(t.TempDir(), "missing")
	if err := h.LoadFromFile(path); err != nil {
		t.Fatal(err)
	}
	if h.Len() != 0 {
		t.Fatalf("len = %d, want 0", h.Len())
	}
}

func TestLineHistory_FileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repl_history")
	h := NewLineHistory()
	h.SetFilePath(path)
	h.Add("line")
	if err := h.WriteFile(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != historyFileMode {
		t.Fatalf("mode = %o, want %o", info.Mode().Perm(), historyFileMode)
	}
}

func TestLineHistory_FileSync(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repl_history")

	h := NewLineHistory()
	h.SetFilePath(path)
	h.Add("first")
	h.Add("second")

	h2 := NewLineHistory()
	if err := h2.LoadFromFile(path); err != nil {
		t.Fatal(err)
	}
	if h2.Len() != 2 || h2.At(0) != "first" || h2.At(1) != "second" {
		t.Fatalf("loaded = [%q %q]", h2.At(0), h2.At(1))
	}

	h.Clear()
	h.SetLine("reloaded")
	if err := h.ReloadFromFile(); err != nil {
		t.Fatal(err)
	}
	if h.Len() != 2 || h.At(0) != "first" {
		t.Fatalf("after -r = [%q %q]", h.At(0), h.At(1))
	}

	h.Add("third")
	if err := h.AppendPendingToFile(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first\nsecond\nthird\n" {
		t.Fatalf("file = %q", string(data))
	}
}
