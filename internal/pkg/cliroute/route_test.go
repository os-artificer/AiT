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

package cliroute

import (
	"testing"
)

func TestRoute_IsSSHConnectArgs(t *testing.T) {
	SetReservedSubcommandNames([]string{"shell", "session", "config"})
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{}, false},
		{[]string{"--help"}, false},
		{[]string{"-h"}, false},
		{[]string{"login", "user@host"}, false},
		{[]string{"shell"}, false},
		{[]string{"user@host"}, true},
		{[]string{"--save"}, true},
		{[]string{"devhost"}, true},
		{[]string{"list"}, false},
		{[]string{"load"}, false},
		{[]string{"suspend"}, false},
		{[]string{"models"}, false},
	}
	for _, tc := range cases {
		if got := IsSSHConnectArgs(tc.args); got != tc.want {
			t.Fatalf("IsSSHConnectArgs(%v)=%v want %v", tc.args, got, tc.want)
		}
	}
}

func TestRoute_IsMetaCLIFirstWord(t *testing.T) {
	SetReservedSubcommandNames([]string{"shell", "session"})
	if !IsMetaCLIFirstWord("list") {
		t.Fatal("list should be meta CLI")
	}
	if !IsMetaCLIFirstWord("shell") {
		t.Fatal("shell should be meta CLI")
	}
	if IsMetaCLIFirstWord("devhost") {
		t.Fatal("devhost should not be meta CLI")
	}
}

func TestRoute_LoginPrefixNotSSH(t *testing.T) {
	SetReservedSubcommandNames([]string{"shell"})
	if IsSSHConnectArgs([]string{"login", "user@host"}) {
		t.Fatal("login prefix must not be SSH connect at top level")
	}
	if !IsRemovedTopLevelWord("login") {
		t.Fatal("login must be a removed top-level word")
	}
}
