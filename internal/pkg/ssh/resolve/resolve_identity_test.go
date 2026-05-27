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

package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/stricthostkey"
)

func TestParseSSHGOutputMultipleIdentityFile(t *testing.T) {
	out := []byte(`identityfile ~/.ssh/id_rsa
identityfile ~/.ssh/id_dsa
hostname example.com
`)
	cfg := parseSSHGOutput(out)
	if len(cfg.multi["identityfile"]) != 2 {
		t.Fatalf("identityfile count=%d want 2", len(cfg.multi["identityfile"]))
	}
	if cfg.multi["identityfile"][0] != "~/.ssh/id_rsa" || cfg.multi["identityfile"][1] != "~/.ssh/id_dsa" {
		t.Fatalf("order: %v", cfg.multi["identityfile"])
	}
}

func TestFirstExistingIdentity(t *testing.T) {
	dir := t.TempDir()
	rsa := filepath.Join(dir, "id_rsa")
	if err := os.WriteFile(rsa, []byte("not-a-real-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "id_dsa")
	got := firstExistingIdentity([]string{missing, rsa})
	if got != rsa {
		t.Fatalf("got %q want %q", got, rsa)
	}
}

func TestNormalizeStrictHostKeyChecking(t *testing.T) {
	cases := map[string]string{
		"": "ask", "ask": "ask", "yes": "yes", "no": "no", "accept-new": "accept-new",
	}
	for in, want := range cases {
		if got := stricthostkey.NormalizeStrictHostKeyChecking(in); got != want {
			t.Errorf("%q -> %q want %q", in, got, want)
		}
	}
}
