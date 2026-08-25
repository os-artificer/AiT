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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestRunSetup_DeclinedInstall_OCONF(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cfg := &config.Config{}
	var out bytes.Buffer
	opts := Options{
		BaseURL:    srv.URL + "/v1",
		ModelsDir:  filepath.Join(t.TempDir(), "models"),
		NoInstall:  true,
		NoPull:     true,
		Out:        &out,
		Err:        &out,
		HTTPClient: srv.Client(),
		Runner:     &fakeRunner{paths: map[string]bool{}},
	}
	res, err := RunSetup(context.Background(), cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !res.ConfigWritten || !res.DeclinedInstall {
		t.Fatalf("res=%+v", res)
	}
	if !cfg.TieredEnabled() {
		t.Fatal("tiered should be enabled")
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestManualSetupInstructions(t *testing.T) {
	cfg := &config.Config{}
	ApplyTieredConfig(cfg, "llama3.2", config.DefaultOllamaBase)
	var buf bytes.Buffer
	ManualSetupInstructions(&buf, cfg)
	out := buf.String()
	for _, sub := range []string{"ollama serve", "ollama pull", toolname.AICLI("ollama setup"), "tiered", "enabled: true"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in:\n%s", sub, out)
		}
	}
}
