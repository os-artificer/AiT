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

package aithost_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aithost"
	"github.com/os-artificer/AiT/internal/aithost/registry"
	"github.com/os-artificer/AiT/internal/aithost/top"
	"github.com/os-artificer/AiT/internal/pkg/hostdash"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

func TestWriteRootHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := aithost.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Commands:", "Examples:",
		"top", "cpu", "mem", "disk", "net", "proc", "snapshot", "sources", "version",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestCLI_HelpFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"--help"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("%s", out.String())
	}
}

func TestCLI_VersionFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"-V"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Version(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"version"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Snapshot(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"snapshot"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"cpu:", "mem:", "disk", "net", "PID"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCLI_SnapshotJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"snapshot", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	var snap hoststats.Snapshot
	if err := json.Unmarshal(out.Bytes(), &snap); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
}

func TestCLI_CPU(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"cpu"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cpu:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_MemJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"mem", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"used_percent"`) && !strings.Contains(out.String(), "used_percent") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Disk(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"disk"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "disk") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Net(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"net"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "net") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Proc(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"proc"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "PID") || !strings.Contains(got, "CPU%") {
		t.Fatalf("%q", got)
	}
	lines := strings.Split(strings.TrimSpace(got), "\n")
	// header + at most 10 data rows
	if len(lines) > 11 {
		t.Fatalf("too many lines: %d\n%s", len(lines), got)
	}
}

func TestCLI_ProcJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"proc", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"proc"`) && !strings.Contains(got, `"pid"`) {
		t.Fatalf("%q", got)
	}
}

func TestCLI_ProcSortLimit(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"proc", "--sort", "mem", "--limit", "5"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 1 || len(lines) > 6 {
		t.Fatalf("lines=%d\n%s", len(lines), out.String())
	}
}

func TestCLI_Sources(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aithost.ExecuteWriters([]string{"sources"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"cpu", "mem", "disk", "net", "proc"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestCLI_TopNonTTY(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := aithost.ExecuteWriters([]string{"top"}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error on non-TTY")
	}
	msg := err.Error()
	if !strings.Contains(msg, "TTY") && !strings.Contains(msg, "snapshot") {
		t.Fatalf("%v", err)
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := aithost.ExecuteWriters([]string{"nope"}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistry_EntriesIncludeTop(t *testing.T) {
	names := map[string]bool{}
	for _, e := range registry.Entries() {
		names[e.Name] = true
	}
	for _, want := range []string{"top", "cpu", "proc", "version"} {
		if !names[want] {
			t.Fatalf("missing registered %s", want)
		}
	}
}

func TestTopModel_IncludesProcPanel(t *testing.T) {
	panels := hostdash.DefaultPanels()
	m := struct {
		panels []hostdash.Panel
	}{panels: panels}
	snap := hoststats.Snapshot{
		CPU:  hoststats.CPUStat{Percent: 10},
		Mem:  hoststats.MemStat{UsedPercent: 20, Used: 1, Total: 2},
		Disk: hoststats.DiskStat{Mount: "/", UsedPercent: 30},
		Net:  hoststats.NetStat{Name: "lo"},
		Processes: []hoststats.ProcStat{
			{PID: 42, Name: "tester", CPUPercent: 9},
		},
	}
	for _, p := range m.panels {
		p.Update(snap)
	}
	view := panels[4].View(100, 12)
	if !strings.Contains(view, "PROC") || !strings.Contains(view, "PID") || !strings.Contains(view, "tester") {
		t.Fatalf("%q", view)
	}
	_ = top.Name
}
