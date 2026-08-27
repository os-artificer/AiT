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

package aissh

import (
	"bytes"
	"io"
	"strings"
	"testing"

	_ "github.com/os-artificer/AiT/internal/ssh/registry/all"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

func TestExecuteWriters_EmptyShowsRootHelp(t *testing.T) {
	var out bytes.Buffer
	err := ExecuteWriters(nil, &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "OpenSSH-compatible flags:") {
		t.Fatalf("root help missing SSH section: %s", out.String())
	}
	if strings.Contains(out.String(), "  login       ") {
		t.Fatalf("root help must not list login subcommand: %s", out.String())
	}
}

func TestExecuteWriters_LoginUnknown(t *testing.T) {
	err := ExecuteWriters([]string{"login", "user@host"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("want unknown command, got %v", err)
	}
}

func TestExecuteWriters_FlagOnlyDestinationRequired(t *testing.T) {
	err := ExecuteWriters([]string{"--save"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected destination error")
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try: %q", err.Error())
	}
}

func TestExecuteWriters_SSHDefaultConnect(t *testing.T) {
	err := ExecuteWriters([]string{"user@example.com"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected dial/resolve error without real host")
	}
	if pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("must route to SSH connect, not Cobra: %v", err)
	}
	if strings.Contains(err.Error(), `unknown command "user@example.com"`) {
		t.Fatalf("must route to SSH connect, not Cobra: %v", err)
	}
}

func TestExecuteWriters_UnknownFlagTryConnectHelp(t *testing.T) {
	err := ExecuteWriters([]string{"-Z", "user@host"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected unknown flag error")
	}
	msg := err.Error()
	if strings.Contains(msg, "login") {
		t.Fatalf("unknown-flag hint must not mention login: %q", msg)
	}
	if !strings.Contains(msg, clihints.TryConnectHelp) {
		t.Fatalf("missing connect help try line: %q", msg)
	}
}

func TestExecuteWriters_VerboseClusterRoutesConnect(t *testing.T) {
	err := ExecuteWriters([]string{"-vvv", "user@example.com"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected dial/resolve error without real host")
	}
	if strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("-vvv must be accepted as ssh-compat no-op: %v", err)
	}
	if pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("must route to SSH connect, not Cobra: %v", err)
	}
}
