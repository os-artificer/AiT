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

import "testing"

// TestDialChain_JumpUsesResolvedHostPort (PJ2) verifies jumpOptionsForHop prefers JumpDialResolver.
func TestDialChain_JumpUsesResolvedHostPort(t *testing.T) {
	prev := JumpDialResolver
	t.Cleanup(func() { JumpDialResolver = prev })

	JumpDialResolver = func(jump string, target Options) (Options, error) {
		if jump != "devhost" {
			t.Fatalf("jump = %q", jump)
		}
		out := target
		out.Host = "jump.example.com"
		out.Port = 2222
		out.User = "root"
		return out, nil
	}

	opts, err := jumpOptionsForHop("devhost", Options{Host: "final.example", Port: 22, User: "root"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Host != "jump.example.com" || opts.Port != 2222 || opts.User != "root" {
		t.Fatalf("opts = %+v", opts)
	}
}
