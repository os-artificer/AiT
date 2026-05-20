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

package ait

import (
	"reflect"
	"testing"
)

func TestParseShellLine(t *testing.T) {
	args, err := parseShellLine(`login --host 192.168.1.1 --user root`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"login", "--host", "192.168.1.1", "--user", "root"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("got %v want %v", args, want)
	}
}

func TestParseLoginArgs(t *testing.T) {
	host, alias, user, _, _, port, tmux, err := parseLoginArgs([]string{"--host", "h", "--alias", "a", "--port", "2222", "--tmux"})
	if err != nil {
		t.Fatal(err)
	}
	if host != "h" || alias != "a" || user != "" || port != 2222 || !tmux {
		t.Errorf("host=%q alias=%q port=%d tmux=%v", host, alias, port, tmux)
	}
}
