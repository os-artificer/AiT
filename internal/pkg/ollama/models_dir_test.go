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

package ollama

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func TestResolveModelsDir_DefaultAitData(t *testing.T) {
	dir, err := ResolveModelsDir(nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dir, string(filepath.Separator)+"ollama"+string(filepath.Separator)+"models") {
		t.Fatalf("unexpected default dir %q", dir)
	}
}

func TestApplyModelsDirEnv(t *testing.T) {
	cmd := exec.Command("true")
	ApplyModelsDirEnv(cmd, "/tmp/ait-ollama-models")
	found := false
	for _, e := range cmd.Env {
		if e == envModelsDir+"=/tmp/ait-ollama-models" {
			found = true
		}
	}
	if !found {
		t.Fatalf("env=%v", cmd.Env)
	}
}

func TestSetProviderModelsDir(t *testing.T) {
	cfg := &config.Config{}
	cfg.AI.Providers = map[string]config.ProviderConfig{}
	dir := filepath.Join(t.TempDir(), "models")
	if err := SetProviderModelsDir(cfg, dir); err != nil {
		t.Fatal(err)
	}
	if got := ModelsDirFromConfig(cfg); got != dir {
		t.Fatalf("got %q want %q", got, dir)
	}
}

func TestChooseModelsDir_AssumeYesPersistsFlagDir(t *testing.T) {
	cfg := &config.Config{}
	cfg.AI.Providers = map[string]config.ProviderConfig{}
	flagDir := filepath.Join(t.TempDir(), "models")
	var out strings.Builder
	dir, err := ChooseModelsDir(strings.NewReader(""), &out, cfg, true, flagDir)
	if err != nil {
		t.Fatal(err)
	}
	if dir != flagDir {
		t.Fatalf("got %q want %q", dir, flagDir)
	}
	if ModelsDirFromConfig(cfg) != flagDir {
		t.Fatalf("config %q", ModelsDirFromConfig(cfg))
	}
}
