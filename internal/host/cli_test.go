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

package aihost_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/host"
	"github.com/os-artificer/AiT/internal/host/registry"
	"github.com/os-artificer/AiT/internal/host/top"
	"github.com/os-artificer/AiT/internal/pkg/hostdash"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

func TestWriteRootHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := aihost.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Commands:", "Examples:",
		"top", "cpu", "mem", "disk", "net", "proto", "proc", "snapshot", "sources", "version",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	assertCommandsVersionLast(t, out, "  top ")
}

func assertCommandsVersionLast(t *testing.T, out, productRow string) {
	t.Helper()
	idx := strings.Index(out, "Commands:")
	if idx < 0 {
		t.Fatal("missing Commands:")
	}
	block := out[idx:]
	prod := strings.Index(block, productRow)
	ver := strings.Index(block, "  version ")
	if prod < 0 || ver < 0 {
		t.Fatalf("missing rows product=%q version in:\n%s", productRow, block)
	}
	if prod > ver {
		t.Fatalf("version must list after product commands: prod@%d ver@%d\n%s", prod, ver, block)
	}
}

func TestCLI_HelpFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"--help"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("%s", out.String())
	}
	if !strings.Contains(out.String(), "aithost") {
		t.Fatalf("direct binary help should use aithost prefix: %s", out.String())
	}
}

func TestCLI_HelpUsesDispatchedPrefix(t *testing.T) {
	t.Setenv("AIT_PLUGIN_CMD", "host")
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"--help"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "host [command]") {
		t.Fatalf("expected host prefix: %s", out.String())
	}
}

func TestCLI_VersionFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"-V"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Version(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"version"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Snapshot(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"snapshot"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"USER:", "OS:", "CPU:", "MEM:", "KERNEL:", "cpu:", "mem:", "disk", "net", "tcp:", "udp:", "arp:", "icmp:", "PID"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "HOST:") {
		t.Fatalf("hostname must be omitted from header:\n%s", got)
	}
}

func TestCLI_SnapshotJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"snapshot", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	var snap hoststats.Snapshot
	if err := json.Unmarshal(out.Bytes(), &snap); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if snap.Host.OS == "" || snap.Host.Arch == "" || snap.Host.CPUs < 1 {
		t.Fatalf("json host: %+v", snap.Host)
	}
	_ = snap.Proto.TCP.Estab
}

func TestCLI_CPU(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"cpu"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "cpu:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_MemJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"mem", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"used_percent"`) && !strings.Contains(out.String(), "used_percent") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Disk(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"disk"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "disk") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Net(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "net") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_NetTrafficNoConnHeader(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Contains(got, "PROTO") || strings.Contains(got, "PROCESS") {
		t.Fatalf("traffic mode must not print conn table: %q", got)
	}
}

func TestCLI_NetJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if _, ok := got["net"]; !ok {
		t.Fatalf("missing net key: %s", out.String())
	}
	if _, ok := got["connections"]; ok {
		t.Fatal("connections key must not appear without --conn")
	}
}

func TestCLI_NetHelp(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net", "--help"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"--conn", "AGE", "CREATED", "Examples:", "nf_conntrack", "delta-time", "TCP LISTEN", "--port", "8080"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCLI_NetConn(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net", "--conn", "--limit", "5"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"PROTO", "STATE", "LOCAL", "REMOTE", "PID", "PROCESS", "AGE", "CREATED"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCLI_NetConnPort(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net", "--conn", "--port", "1", "--limit", "5"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "CREATED") {
		t.Fatalf("missing CREATED header:\n%s", got)
	}
}

func TestCLI_NetConnJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"net", "--conn", "--json", "--limit", "5"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["connections"]; !ok {
		t.Fatalf("%s", out.String())
	}
}

func TestCLI_NetFiltersRequireConn(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := aihost.ExecuteWriters([]string{"net", "--pid", "1"}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--conn") {
		t.Fatalf("%v", err)
	}
}

func TestCLI_Proto(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"proto"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"tcp:", "udp:", "arp:", "icmp:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCLI_ProtoJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"proto", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{`"tcp"`, `"udp"`, `"arp"`, `"icmp"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestCLI_Proc(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"proc"}, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "PID") || !strings.Contains(got, "CPU%") {
		t.Fatalf("%q", got)
	}
	if !strings.Contains(got, "PEER") {
		t.Fatalf("missing PEER header:\n%s", got)
	}
	if !strings.Contains(got, "LOG") {
		t.Fatalf("missing LOG header:\n%s", got)
	}
	if strings.Contains(got, "NET/s") {
		t.Fatalf("NET/s should be replaced by PEER:\n%s", got)
	}
	lines := strings.Split(strings.TrimSpace(got), "\n")
	// header + at most 10 data rows
	if len(lines) > 11 {
		t.Fatalf("too many lines: %d\n%s", len(lines), got)
	}
}

func TestCLI_ProcJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"proc", "--json"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"proc"`) && !strings.Contains(got, `"pid"`) {
		t.Fatalf("%q", got)
	}
}

func TestCLI_ProcSortLimit(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"proc", "--sort", "mem", "--limit", "5"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 1 || len(lines) > 6 {
		t.Fatalf("lines=%d\n%s", len(lines), out.String())
	}
}

func TestCLI_Sources(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"sources"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"cpu", "mem", "disk", "net", "proto", "proc"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestCLI_TopNonTTY(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := aihost.ExecuteWriters([]string{"top"}, &out, &errBuf)
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
	err := aihost.ExecuteWriters([]string{"nope"}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCLI_PluginManifest(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := aihost.ExecuteWriters([]string{"ait-plugin-manifest"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"binary": "aithost"`) {
		t.Fatalf("manifest = %q", out.String())
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
	view := panels[hostdash.PanelPROC].View(100, 12)
	if !strings.Contains(view, "PROC") || !strings.Contains(view, "PID") || !strings.Contains(view, "tester") {
		t.Fatalf("%q", view)
	}
	_ = top.Name
}
