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
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	sshparse "github.com/os-artificer/AiT/internal/pkg/ssh/parse"
)

func TestResolveGithub(t *testing.T) {
	if _, err := platform.OpenSSHPath(); err != nil {
		t.Skip("ssh not installed:", err)
	}
	spec, err := sshparse.Parse([]string{"github.com"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := Resolve(spec)
	if err != nil {
		t.Fatal(err)
	}
	if r.Host == "" {
		t.Fatal("empty host")
	}
	if r.Port == 0 {
		t.Fatal("empty port")
	}
}

func TestClientOptions_LoginDestination(t *testing.T) {
	r := Resolved{
		Host:        "dev.ylgeeker.com",
		Destination: "devhost",
		Port:        22,
		User:        "root",
	}
	opts := r.ClientOptions()
	if opts.LoginDestination != "devhost" {
		t.Fatalf("LoginDestination = %q, want devhost", opts.LoginDestination)
	}
	if opts.Host != "dev.ylgeeker.com" {
		t.Fatalf("Host = %q", opts.Host)
	}
}
