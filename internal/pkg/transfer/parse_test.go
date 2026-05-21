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

package transfer

import "testing"

func TestParseRemoteDest(t *testing.T) {
	user, host, path, err := ParseRemoteDest("user@host.example:/var/log/app.log")
	if err != nil || user != "user" || host != "host.example" || path != "/var/log/app.log" {
		t.Fatalf("got user=%q host=%q path=%q err=%v", user, host, path, err)
	}
	user, host, path, err = ParseRemoteDest("user@[2001:db8::1]:/tmp/x")
	if err != nil || user != "user" || host != "2001:db8::1" || path != "/tmp/x" {
		t.Fatalf("ipv6: user=%q host=%q path=%q err=%v", user, host, path, err)
	}
}
