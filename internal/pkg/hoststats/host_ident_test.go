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

package hoststats

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFormatHostHeader_IncludesUserAndHost(t *testing.T) {
	got := FormatHostHeader(HostIdent{
		Hostname: "devhost.example.com",
		User:     "alice",
		OS:       "darwin",
		Arch:     "arm64",
		Kernel:   "24.6.0",
		CPUs:     8,
	}, 1024<<20, time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC))
	for _, want := range []string{
		"USER: alice",
		"OS: darwin/arm64",
		"CPU: 8",
		"MEM: 1024 MB",
		"KERNEL: 24.6.0",
		"2026/08/16 10:40:38",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "HOST:") || strings.Contains(got, "devhost.example.com") {
		t.Fatalf("hostname must be omitted: %q", got)
	}
}

func TestFormatHostHeader_SkipsEmpty(t *testing.T) {
	got := FormatHostHeader(HostIdent{OS: "linux"}, 0, time.Time{})
	if got != "OS: linux" {
		t.Fatalf("got %q", got)
	}
}

func TestWrapHostHeader_Narrow(t *testing.T) {
	groups := HostHeaderGroups(HostIdent{
		User:   "alice",
		OS:     "darwin",
		Arch:   "arm64",
		Kernel: "24.6.0",
		CPUs:   8,
	}, 1024<<20, time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC))
	oneLine := strings.Join(groups, "  ")
	lines := WrapHostHeader(groups, 40)
	if len(lines) < 2 {
		t.Fatalf("expected wrap at width 40, got %d lines from %q", len(lines), oneLine)
	}
	for i, line := range lines {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("line %d not left-aligned: %q", i, line)
		}
		if strings.Contains(line, "USER:") && !strings.Contains(line, "USER: alice") {
			t.Fatalf("USER group split: %q", line)
		}
	}
}

func TestCollectHostIdent_HasOSArch(t *testing.T) {
	ident := CollectHostIdent()
	if ident.OS == "" || ident.Arch == "" || ident.CPUs < 1 {
		t.Fatalf("incomplete ident: %+v", ident)
	}
}

func TestSample_FillsHost(t *testing.T) {
	c := NewCollector()
	snap, err := c.Sample(context.Background(), ProcessQuery{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Host.OS == "" || snap.Host.Arch == "" || snap.Host.CPUs < 1 {
		t.Fatalf("sample host: %+v", snap.Host)
	}
}

func TestSample_FillsProto(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()
	if _, err := c.Sample(ctx, ProcessQuery{Limit: 1}); err != nil {
		t.Fatal(err)
	}
	snap, err := c.Sample(ctx, ProcessQuery{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	// Soft-fail platforms may leave rates at 0; Proto must still round-trip.
	p := snap.Proto
	_ = p.TCP.InPerSec + p.UDP.OutPerSec + p.ARP.LookupsPerSec + p.ICMP.OutPerSec
	_ = p.TCP.Estab + p.UDP.NoPort + p.ARP.Entries + p.ICMP.InErr
}

func TestDefaultMount_Platform(t *testing.T) {
	got := defaultMount()
	if runtime.GOOS == "windows" {
		if got != `C:\` {
			t.Fatalf("windows mount %q", got)
		}
		return
	}
	if got != "/" {
		t.Fatalf("unix mount %q", got)
	}
}

func TestSample_DiskMountMatchesPlatform(t *testing.T) {
	c := NewCollector()
	snap, err := c.Sample(context.Background(), ProcessQuery{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Disk.Mount != defaultMount() {
		t.Fatalf("disk mount %q want %q", snap.Disk.Mount, defaultMount())
	}
}

func TestProtoStatsFromRaw_Rates(t *testing.T) {
	prev := protoRaw{tcpIn: 10, tcpOut: 20, tcpRetrans: 1, udpIn: 5, arpLookups: 2}
	cur := protoRaw{tcpIn: 30, tcpOut: 50, tcpRetrans: 3, tcpEstab: 4, tcpOpens: 7,
		udpIn: 15, udpNoPort: 1, arpLookups: 12, arpEnts: 3, arpIncomp: 1}
	got := protoStatsFromRaw(cur, prev, 2, true)
	if got.TCP.InPerSec != 10 || got.TCP.OutPerSec != 15 || got.TCP.RetransPerSec != 1 {
		t.Fatalf("tcp rates: %+v", got.TCP)
	}
	if got.TCP.Estab != 4 || got.TCP.Opens != 7 {
		t.Fatalf("tcp gauges: %+v", got.TCP)
	}
	if got.UDP.InPerSec != 5 || got.UDP.NoPort != 1 {
		t.Fatalf("udp: %+v", got.UDP)
	}
	if got.ARP.LookupsPerSec != 5 || got.ARP.Entries != 3 || got.ARP.Incomplete != 1 {
		t.Fatalf("arp: %+v", got.ARP)
	}
}

func TestFormatPacketRate(t *testing.T) {
	if got := FormatPacketRate(42); got != "42" {
		t.Fatalf("%q", got)
	}
	if got := FormatPacketRate(1500); got != "1.5K" {
		t.Fatalf("%q", got)
	}
}
