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

package net

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWriteHelp_IncludesConn(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"--conn", "AGE", "CREATED", "Usage:", "Examples:", "net --conn",
		"TCP LISTEN", "UDP", "--port", "8080",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRunConnections_TableHeaders(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--conn", "--limit", "5"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"PROTO", "STATE", "LOCAL", "REMOTE", "PID", "PROCESS", "AGE", "CREATED"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(got), "net ") {
		t.Fatalf("traffic line leaked: %q", got)
	}
}

func TestRunConnections_JSONEnvelope(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--conn", "--json", "--limit", "5"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if _, ok := got["connections"]; !ok {
		t.Fatalf("missing connections: %s", out.String())
	}
	if _, ok := got["net"]; ok {
		t.Fatal("traffic net key must not appear")
	}
}

func TestRunConnections_ListenFilter(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--conn", "--listen", "--limit", "20"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "PROTO") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] != "LISTEN" {
			t.Fatalf("non-listen row: %q", line)
		}
	}
}

func TestRunConnections_ProtoUDP(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--conn", "--proto", "udp", "--json", "--limit", "20"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var env struct {
		Connections []struct {
			Proto string `json:"proto"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	for _, c := range env.Connections {
		if c.Proto != "udp" {
			t.Fatalf("proto=%q", c.Proto)
		}
	}
}

func TestValidate_FiltersRequireConn(t *testing.T) {
	cmd := New()
	cmd.SetArgs([]string{"--limit", "5"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--conn") {
		t.Fatalf("%v", err)
	}
}

func TestValidate_ListenStateExclusive(t *testing.T) {
	cmd := New()
	cmd.SetArgs([]string{"--conn", "--listen", "--state", "ESTABLISHED"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "mutually exclusive") && !strings.Contains(msg, "--listen") {
		t.Fatalf("%v", err)
	}
}

func TestValidate_BadPort(t *testing.T) {
	cmd := New()
	cmd.SetArgs([]string{"--conn", "--port", "0"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate_BadLimit(t *testing.T) {
	for _, args := range [][]string{
		{"--conn", "--limit", "0"},
		{"--conn", "--limit", "-1"},
	} {
		cmd := New()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}

func TestRunConnections_WatchHumanSeparator(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--conn", "--watch", "--interval", "1s", "--limit", "3"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "---") {
		t.Fatalf("want human watch separator: %q", out.String())
	}
}

func TestRunConnections_WatchJSONNoSeparator(t *testing.T) {
	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--conn", "--watch", "--json", "--interval", "1s", "--limit", "3"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "---") {
		t.Fatalf("JSON watch must not use ---:\n%s", out.String())
	}
	if !strings.Contains(out.String(), `"connections"`) {
		t.Fatalf("%s", out.String())
	}
}
