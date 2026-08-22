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

package ait

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewAITHistory_Basename(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIT_HOME", dir)

	h := newAITHistory(io.Discard)
	path := h.FilePath()
	if path == "" || !strings.HasSuffix(path, historyBasename) {
		t.Fatalf("FilePath = %q, want suffix %q", path, historyBasename)
	}

	h.Add("version")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("history file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "repl_history")); err == nil {
		t.Fatal("must not use repl_history")
	}
}
