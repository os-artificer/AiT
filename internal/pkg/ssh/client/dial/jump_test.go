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

package dial

import "testing"

func TestParseJumpDestination_UserHostPort(t *testing.T) {
	u, h, p, err := ParseJumpDestination("alice@bastion.example.com:2222")
	if err != nil {
		t.Fatal(err)
	}
	if u != "alice" || h != "bastion.example.com" || p != 2222 {
		t.Errorf("got u=%q h=%q p=%d", u, h, p)
	}
}

func TestParseJumpDestination_IPv6(t *testing.T) {
	u, h, p, err := ParseJumpDestination("[fd00::1]:2200")
	if err != nil {
		t.Fatal(err)
	}
	if u != "" || h != "fd00::1" || p != 2200 {
		t.Errorf("got u=%q h=%q p=%d", u, h, p)
	}
}

func TestParseJumpDestination_HostOnly(t *testing.T) {
	u, h, p, err := ParseJumpDestination("host")
	if err != nil {
		t.Fatal(err)
	}
	if u != "" || h != "host" || p != 0 {
		t.Errorf("got u=%q h=%q p=%d", u, h, p)
	}
}

func TestParsePortString(t *testing.T) {
	p, err := ParsePortString("22")
	if err != nil || p != 22 {
		t.Fatalf("p=%d err=%v", p, err)
	}
	if _, err := ParsePortString("abc"); err == nil {
		t.Error("expected error for non-numeric port")
	}
}
