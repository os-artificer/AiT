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

import (
	"reflect"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
)

func TestArgsFromMetaLine_Get(t *testing.T) {
	cmd, args, ok := ArgsFromMetaLine("!ai-ssh get /etc/hosts ./hosts")
	if !ok || cmd != "get" {
		t.Fatalf("cmd=%q ok=%v", cmd, ok)
	}
	if !reflect.DeepEqual(args, []string{"/etc/hosts", "./hosts"}) {
		t.Fatalf("args=%v", args)
	}
}

func TestArgsFromMetaLine_Put(t *testing.T) {
	cmd, args, ok := ArgsFromMetaLine("!ai-ssh put ./a /b")
	if !ok || cmd != "put" {
		t.Fatalf("cmd=%q ok=%v", cmd, ok)
	}
	if !reflect.DeepEqual(args, []string{"./a", "/b"}) {
		t.Fatalf("args=%v", args)
	}
}

func TestArgsFromMetaLine_NonTransfer(t *testing.T) {
	if _, _, ok := ArgsFromMetaLine("!ai-ssh list"); ok {
		t.Fatal("non-transfer meta should not match")
	}
	if _, _, ok := ArgsFromMetaLine("ai-ssh get x y"); ok {
		t.Fatal("missing prefix should not match")
	}
}

// TestArgsFromMetaLine_UsesMetaPrefix (C21) keeps argv parsing aligned with meta.PrefixAitSpace.
func TestArgsFromMetaLine_UsesMetaPrefix(t *testing.T) {
	line := meta.PrefixAitSpace + "get /r ./l"
	cmd, args, ok := ArgsFromMetaLine(line)
	if !ok || cmd != "get" {
		t.Fatalf("cmd=%q ok=%v", cmd, ok)
	}
	if !reflect.DeepEqual(args, []string{"/r", "./l"}) {
		t.Fatalf("args=%v", args)
	}
}

func TestUsageHintErr_Format(t *testing.T) {
	err := UsageHintErr(UsageGet)
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, UsageGet) {
		t.Fatalf("missing usage sentinel: %q", s)
	}
	if !strings.Contains(s, "!ai-ssh get <remote-path> <local-path>") {
		t.Fatalf("missing get try line: %q", s)
	}
}
