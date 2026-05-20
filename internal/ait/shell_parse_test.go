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

	"github.com/os-artificer/AiT/internal/pkg/sshargs"
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

func TestParseLoginArgsLegacy(t *testing.T) {
	spec, err := sshargs.Parse([]string{"--host", "h", "--alias", "a", "--port", "2222", "--tmux"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Destination != "h" || spec.Alias != "a" || spec.LegacyPort != 2222 || !spec.Tmux {
		t.Fatalf("spec=%+v", spec)
	}
}
