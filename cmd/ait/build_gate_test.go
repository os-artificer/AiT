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

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const externalXiwhZmodem = "github.com/xiwh/zmodem"
const externalGoOpenAI = "github.com/sashabaranov/go-openai"

// TestGoMod_NoExternalXiwhZmodem (M1) ensures zmodem is vendored under pkg/third-party only.
func TestGoMod_NoExternalXiwhZmodem(t *testing.T) {
	modPath := filepath.Join("..", "..", "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, externalXiwhZmodem) {
		t.Fatalf("go.mod must not require external %s; use pkg/third-party", externalXiwhZmodem)
	}
}

// TestGoMod_NoGoOpenAI (E6) ensures AI routing no longer depends on sashabaranov/go-openai.
func TestGoMod_NoGoOpenAI(t *testing.T) {
	modPath := filepath.Join("..", "..", "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), externalGoOpenAI) {
		t.Fatalf("go.mod must not require %s; use cloudwego/eino", externalGoOpenAI)
	}
}

// TestBuild_FreshModCache (M1) builds cmd/ait with an empty module cache when not -short.
func TestBuild_FreshModCache(t *testing.T) {
	if testing.Short() {
		t.Skip("fresh mod cache build skipped in -short")
	}

	tmpBase, err := filepath.Abs(filepath.Join("..", "..", ".test-fresh-gomodcache"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpBase, 0o700); err != nil {
		t.Fatal(err)
	}
	tmpCache, err := os.MkdirTemp(tmpBase, "cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpCache)

	cmd := exec.Command("go", "build", "-o", filepath.Join(t.TempDir(), "ait"), ".")
	cmd.Env = append(os.Environ(), "GOMODCACHE="+tmpCache)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build with fresh GOMODCACHE: %v\n%s", err, out)
	}
}
