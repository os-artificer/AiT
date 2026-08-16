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

package top

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hostdash"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestPanelRowHeights_FitsTerminal(t *testing.T) {
	for _, h := range []int{5, 10, 16, 24, 40, 80} {
		for _, headerH := range []int{1, 2, 3} {
			topH, midH, protoH, procH := panelRowHeights(h, headerH)
			sum := topH + midH + protoH + procH + headerH + footerReserve
			if h >= headerH+footerReserve+minPanelRows && sum != h {
				t.Fatalf("h=%d headerH=%d top=%d mid=%d proto=%d proc=%d sum=%d", h, headerH, topH, midH, protoH, procH, sum)
			}
			if topH < 1 || midH < 1 || protoH < 1 || procH < 1 {
				t.Fatalf("h=%d headerH=%d got non-positive row heights", h, headerH)
			}
		}
	}
}

func TestView_NoVerticalOverflow(t *testing.T) {
	m := newModel(context.Background(), 0, hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{}))
	m.ready = true
	m.host = hoststats.HostIdent{User: "alice", OS: "darwin", Arch: "arm64", Kernel: "24.6.0", CPUs: 8}
	m.memTotal = 1024 << 20
	m.clock = time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC)
	for _, h := range []int{16, 24, 40} {
		m.width = 80
		m.height = h
		for _, p := range m.panels {
			p.Update(hoststats.Snapshot{
				CPU:       hoststats.CPUStat{Percent: 10},
				Mem:       hoststats.MemStat{UsedPercent: 50, Used: 1, Total: m.memTotal},
				Processes: []hoststats.ProcStat{{PID: 1, Name: "init"}},
			})
		}
		view := m.View()
		if got := lipgloss.Height(view); got > h {
			t.Fatalf("h=%d viewHeight=%d > term\n%q", h, got, view)
		}
	}
}

func TestProcSampleLimit_AllByDefault(t *testing.T) {
	if got := procSampleLimit(0); got != 0 {
		t.Fatalf("default top sample is unlimited (0): got %d", got)
	}
	capped := procSampleLimit(5)
	if capped != 5 {
		t.Fatalf("user --limit 5: got %d", capped)
	}
	if got := procSampleLimit(hoststats.MaxProcessLimit + 10); got != hoststats.MaxProcessLimit {
		t.Fatalf("cap at MaxProcessLimit: got %d", got)
	}
}

func TestView_ProcFillsTallWindow(t *testing.T) {
	m := newModel(context.Background(), 0, hoststats.ProcessQuery{Sort: hoststats.SortCPU})
	m.ready = true
	m.width = 80
	m.height = 48
	m.host = hoststats.HostIdent{User: "alice", OS: "linux", Arch: "amd64", Kernel: "6.1", CPUs: 4}
	m.memTotal = 1024 << 20
	m.clock = time.Date(2026, 8, 16, 13, 0, 0, 0, time.UTC)

	procs := make([]hoststats.ProcStat, 40)
	for i := range procs {
		procs[i] = hoststats.ProcStat{PID: int32(200 + i), Name: fmt.Sprintf("job%02d", i+1)}
	}
	snap := hoststats.Snapshot{
		CPU:       hoststats.CPUStat{Percent: 10},
		Mem:       hoststats.MemStat{UsedPercent: 50, Used: 1, Total: m.memTotal},
		Processes: procs,
	}
	for _, p := range m.panels {
		p.Update(snap)
	}

	plain := ansiEscape.ReplaceAllString(m.View(), "")
	_, headerH := renderHostHeader(m.host, m.memTotal, m.clock, m.width)
	_, _, _, procH := panelRowHeights(m.height, headerH)
	want := hostdash.ProcPanelRowCapacity(procH)
	if want <= hoststats.DefaultProcessLimit {
		t.Fatalf("expected capacity > 10 at h=48, got %d", want)
	}
	if !strings.Contains(plain, fmt.Sprintf("job%02d", want)) {
		t.Fatalf("missing last visible job%02d:\n%s", want, plain)
	}
	if strings.Contains(plain, fmt.Sprintf("job%02d", want+1)) {
		t.Fatalf("showed extra row job%02d:\n%s", want+1, plain)
	}
}

func TestView_IncludesHostHeader(t *testing.T) {
	m := newModel(context.Background(), 0, hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{}))
	m.ready = true
	m.width = 80
	m.height = 24
	m.host = hoststats.HostIdent{
		Hostname: "devhost.example.com",
		User:     "alice",
		OS:       "darwin",
		Arch:     "arm64",
		Kernel:   "24.6.0",
		CPUs:     8,
	}
	m.memTotal = 1024 << 20
	m.clock = time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC)
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"USER: alice", "OS: darwin/arm64", "CPU: 8", "MEM: 1024 MB", "KERNEL: 24.6.0"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in:\n%s", want, view)
		}
	}
	if strings.Contains(view, "HOST:") || strings.Contains(view, "devhost.example.com") {
		t.Fatalf("hostname must be omitted:\n%s", view)
	}
}

func TestView_FooterSearchHint(t *testing.T) {
	m := newModel(context.Background(), time.Second, hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{}))
	m.ready = true
	m.width = 100
	m.height = 24
	view := ansiEscape.ReplaceAllString(m.View(), "")
	for _, want := range []string{"/ search", "enter detail", "esc back", "x kill"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in footer:\n%s", want, view)
		}
	}
}

func TestView_HeaderWrapsNarrow(t *testing.T) {
	m := newModel(context.Background(), 0, hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{}))
	m.ready = true
	m.width = 40
	m.height = 24
	m.host = hoststats.HostIdent{
		User:   "alice",
		OS:     "darwin",
		Arch:   "arm64",
		Kernel: "24.6.0",
		CPUs:   8,
	}
	m.memTotal = 1024 << 20
	m.clock = time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC)
	for _, p := range m.panels {
		p.Update(hoststats.Snapshot{
			CPU:       hoststats.CPUStat{Percent: 10},
			Mem:       hoststats.MemStat{Total: m.memTotal},
			Processes: []hoststats.ProcStat{{PID: 1, Name: "init"}},
		})
	}
	view := m.View()
	plain := ansiEscape.ReplaceAllString(view, "")
	first := strings.Split(plain, "\n")[0]
	if strings.HasPrefix(first, " ") {
		t.Fatalf("header not left-aligned: %q", first)
	}
	if !strings.Contains(plain, "USER: alice") {
		t.Fatalf("missing USER in:\n%s", plain)
	}
	if got := lipgloss.Height(view); got > m.height {
		t.Fatalf("viewHeight=%d > term %d", got, m.height)
	}
	_, lines := renderHostHeader(m.host, m.memTotal, m.clock, m.width)
	if lines < 2 {
		t.Fatalf("expected wrap at width 40, got %d lines", lines)
	}
}

func TestRenderHostHeader_ClockRight(t *testing.T) {
	ident := hoststats.HostIdent{
		User:   "alice",
		OS:     "darwin",
		Arch:   "arm64",
		Kernel: "24.6.0",
		CPUs:   8,
	}
	clock := time.Date(2026, 8, 16, 10, 40, 38, 0, time.UTC)
	text, lines := renderHostHeader(ident, 1024<<20, clock, 100)
	plain := ansiEscape.ReplaceAllString(text, "")
	if lines != 1 {
		t.Fatalf("expected 1 line at width 100, got %d: %q", lines, plain)
	}
	if !strings.Contains(plain, "USER: alice") {
		t.Fatalf("missing USER: %q", plain)
	}
	if !strings.HasSuffix(strings.TrimRight(plain, "\n"), "2026/08/16 10:40:38") {
		t.Fatalf("clock not at end: %q", plain)
	}
	// Clock should be right-padded away from the last fact group.
	if !strings.Contains(plain, "KERNEL: 24.6.0") {
		t.Fatalf("missing KERNEL: %q", plain)
	}
	idx := strings.Index(plain, "KERNEL: 24.6.0")
	clockIdx := strings.LastIndex(plain, "2026/08/16 10:40:38")
	if clockIdx <= idx+len("KERNEL: 24.6.0")+2 {
		t.Fatalf("expected padding before clock: %q", plain)
	}
}
