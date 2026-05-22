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
	"testing"
	"time"
)

func TestNormalizeDialOptions(t *testing.T) {
	opts := Options{Port: 0}
	if err := normalizeDialOptions(&opts); err != nil {
		t.Fatal(err)
	}
	if opts.Port != 22 {
		t.Fatalf("port=%d want 22", opts.Port)
	}
	if opts.User == "" {
		t.Fatal("expected default user from user.Current")
	}
}

func TestBuildSSHClientConfigUsesConnectTimeout(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	opts := Options{
		Host: "127.0.0.1", Port: 22, User: "u", Password: "p",
		ConnectTimeout: 3 * time.Second,
	}
	if err := normalizeDialOptions(&opts); err != nil {
		t.Fatal(err)
	}

	cfg, agentConn, err := buildSSHClientConfig(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer abortSSHSetup(agentConn)

	if cfg.Timeout != 3*time.Second {
		t.Fatalf("timeout=%v want 3s", cfg.Timeout)
	}
}
