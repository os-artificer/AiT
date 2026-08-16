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

package hostdash

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultPanels_IncludesProc(t *testing.T) {
	panels := DefaultPanels()
	found := false
	for _, p := range panels {
		if p.Title() == "PROC" {
			found = true
			p.Update(hoststats.Snapshot{
				Processes: []hoststats.ProcStat{{PID: 1, Name: "init", CPUPercent: 1}},
			})
			view := p.View(80, 12)
			if !strings.Contains(view, "PROC") || !strings.Contains(view, "PID") {
				t.Fatalf("view=%q", view)
			}
			if !strings.Contains(view, "init") {
				t.Fatalf("missing process row: %q", view)
			}
		}
	}
	if !found {
		t.Fatal("ProcPanel missing")
	}
}

func TestDefaultPanels_IncludesProto(t *testing.T) {
	panels := DefaultPanels()
	want := []string{"TCP", "UDP", "ARP", "ICMP"}
	found := map[string]bool{}
	for _, p := range panels {
		found[p.Title()] = true
	}
	for _, name := range want {
		if !found[name] {
			t.Fatalf("missing panel %s", name)
		}
	}
	if len(panels) != 9 {
		t.Fatalf("panels=%d want 9", len(panels))
	}
	if panels[PanelPROC].Title() != "PROC" {
		t.Fatalf("PanelPROC index wrong: %s", panels[PanelPROC].Title())
	}
}

func TestProtoPanel_Details(t *testing.T) {
	now := time.Date(2026, 8, 16, 14, 0, 0, 0, time.UTC)
	snap := hoststats.Snapshot{
		Time: now,
		Proto: hoststats.ProtoStats{
			TCP:  hoststats.TCPStat{InPerSec: 10, OutPerSec: 20, Estab: 5, RetransPerSec: 1.5, Opens: 9},
			UDP:  hoststats.UDPStat{InPerSec: 3, OutPerSec: 4, NoPort: 2, InErr: 1},
			ARP:  hoststats.ARPStat{InPerSec: 1, OutPerSec: 2, Entries: 8, Incomplete: 3, LookupsPerSec: 7},
			ICMP: hoststats.ICMPStat{InPerSec: 0, OutPerSec: 1, InErr: 4, OutErr: 6},
		},
	}
	cases := []struct {
		panel *ProtoPanel
		title string
		bits  []string
	}{
		{NewTCPPanel(), "TCP", []string{"TCP", "estab 5", "retrans", "opens 9"}},
		{NewUDPPanel(), "UDP", []string{"UDP", "noport 2", "inerr 1"}},
		{NewARPPanel(), "ARP", []string{"ARP", "ents 8", "incomp 3", "lookup"}},
		{NewICMPPanel(), "ICMP", []string{"ICMP", "inerr 4", "outerr 6"}},
	}
	for _, tc := range cases {
		tc.panel.Update(snap)
		view := tc.panel.View(48, 12)
		plain := stripANSI(view)
		for _, want := range tc.bits {
			if !strings.Contains(plain, want) {
				t.Fatalf("%s missing %q in:\n%s", tc.title, want, plain)
			}
		}
		if !strings.Contains(view, ansiIn) || !strings.Contains(view, ansiOut) {
			t.Fatalf("%s in/out colors missing:\n%s", tc.title, view)
		}
	}
}

func TestProtoPanel_InOutCurves(t *testing.T) {
	p := NewTCPPanel()
	now := time.Date(2026, 8, 16, 14, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		p.Update(hoststats.Snapshot{
			Time: now.Add(time.Duration(i) * time.Second),
			Proto: hoststats.ProtoStats{
				TCP: hoststats.TCPStat{InPerSec: float64(i+1) * 100, OutPerSec: float64(i + 1)},
			},
		})
	}
	if len(p.histIn) != 4 || len(p.histOut) != 4 {
		t.Fatalf("histIn=%d histOut=%d", len(p.histIn), len(p.histOut))
	}
	view := p.View(56, 12)
	if !strings.Contains(view, ansiIn) || !strings.Contains(view, ansiOut) {
		t.Fatalf("TCP dual colors:\n%s", view)
	}
}

func TestProcPanel_FillsAvailableRows(t *testing.T) {
	p := NewProcPanel()
	rows := make([]hoststats.ProcStat, 40)
	for i := range rows {
		rows[i] = hoststats.ProcStat{PID: int32(i + 1), Name: fmt.Sprintf("job%02d", i+1)}
	}
	p.Update(hoststats.Snapshot{Processes: rows})

	outerH := 16
	want := ProcPanelRowCapacity(outerH)
	if want < 2 {
		t.Fatalf("capacity=%d", want)
	}
	plain := stripANSI(p.View(80, outerH))
	if !strings.Contains(plain, fmt.Sprintf("job%02d", want)) {
		t.Fatalf("missing last visible job%02d:\n%s", want, plain)
	}
	if strings.Contains(plain, fmt.Sprintf("job%02d", want+1)) {
		t.Fatalf("showed extra row job%02d:\n%s", want+1, plain)
	}
}

func TestProcPanel_SearchFilter(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 1, Name: "init"},
		{PID: 42, Name: "sshd", Cmdline: "/usr/sbin/sshd -D"},
		{PID: 99, Name: "bash"},
	}})
	if !p.HandleKey("/") {
		t.Fatal("slash should enter search")
	}
	if !p.HandleKey("s") || !p.HandleKey("s") {
		t.Fatal("typing should be consumed")
	}
	plain := stripANSI(p.View(80, 12))
	if !strings.Contains(plain, "PROC /ss") {
		t.Fatalf("search title: %s", plain)
	}
	if !strings.Contains(plain, "sshd") || strings.Contains(plain, "bash") {
		t.Fatalf("filter wrong:\n%s", plain)
	}
	if !p.HandleKey("esc") {
		t.Fatal("esc clears query")
	}
	if p.query != "" {
		t.Fatalf("query=%q", p.query)
	}
	if !p.HandleKey("esc") {
		t.Fatal("esc leaves search")
	}
	if p.mode != procModeList {
		t.Fatalf("mode=%d", p.mode)
	}
	list := stripANSI(p.View(80, 12))
	if !strings.Contains(list, "bash") || strings.Contains(list, "PROC /") {
		t.Fatalf("list mode:\n%s", list)
	}
}

func TestProcPanel_DetailTree(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 1, PPID: 0, Name: "init", CPUPercent: 0.1},
		{PID: 10, PPID: 1, Name: "bash", CPUPercent: 1, Cmdline: "/bin/bash -l", MemPercent: 2},
		{PID: 20, PPID: 10, Name: "vim", CPUPercent: 5, DiskBytesPerSec: 100},
		{PID: 21, PPID: 10, Name: "sleep", CPUPercent: 0},
	}})
	p.HandleKey("/")
	p.HandleKey("b")
	p.HandleKey("a")
	p.HandleKey("enter")
	if p.mode != procModeDetail {
		t.Fatalf("mode=%d", p.mode)
	}
	plain := stripANSI(p.View(80, 20))
	for _, want := range []string{
		"PROC detail", "bash", "/bin/bash", "peer:", "logs:", "ancestors:", "1 init", "10 bash",
		"children:", "20 vim", "21 sleep", "cpu", "mem", "x kill",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("missing %q in:\n%s", want, plain)
		}
	}
}

func TestProcPanel_DetailLogs(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "bash", Logs: []hoststats.ProcLog{
			{Time: time.Date(2026, 8, 16, 16, 15, 44, 0, time.UTC), Message: "session opened"},
			{Time: time.Date(2026, 8, 16, 16, 15, 45, 0, time.UTC), Message: "accepted key"},
		}, LogLine: "accepted key"},
	}})
	p.HandleKey("/")
	p.HandleKey("b")
	p.HandleKey("enter")
	plain := stripANSI(p.View(80, 24))
	if !strings.Contains(plain, "logs:") || !strings.Contains(plain, "session opened") {
		t.Fatalf("detail logs:\n%s", plain)
	}
	if !strings.Contains(plain, "accepted key") {
		t.Fatalf("missing newest log:\n%s", plain)
	}
	p.HandleKey("esc")
	list := stripANSI(p.View(120, 12))
	if !strings.Contains(list, "LOG") || !strings.Contains(list, "accepted") {
		t.Fatalf("list LOG column:\n%s", list)
	}
}

func TestProcPanel_DetailPeers(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "curl", Remotes: []string{"104.18.18.125:443", "127.0.0.1:7890"}},
	}})
	p.HandleKey("/")
	p.HandleKey("c")
	p.HandleKey("enter")
	plain := stripANSI(p.View(80, 24))
	if !strings.Contains(plain, "peer:") {
		t.Fatalf("missing peer section:\n%s", plain)
	}
	if strings.Contains(plain, "disk") && strings.Contains(plain, "peer 104") {
		t.Fatalf("peer should not share the metrics line:\n%s", plain)
	}
	if !strings.Contains(plain, "104.18.18.125:443") || !strings.Contains(plain, "127.0.0.1:7890") {
		t.Fatalf("full peer records:\n%s", plain)
	}
	idxA := strings.Index(plain, "104.18.18.125:443")
	idxB := strings.Index(plain, "127.0.0.1:7890")
	if idxA < 0 || idxB < 0 || !strings.Contains(plain[idxA:idxB], "\n") {
		t.Fatalf("peers should be one per line:\n%s", plain)
	}
}

func TestProcPanel_DetailPeers_TruncatesToWindow(t *testing.T) {
	remotes := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		remotes = append(remotes, fmt.Sprintf("192.0.2.%d:443", i+1))
	}
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "curl", Remotes: remotes},
	}})
	p.HandleKey("/")
	p.HandleKey("c")
	p.HandleKey("enter")
	plain := stripANSI(p.View(80, 10))
	if !strings.Contains(plain, "192.0.2.1:443") {
		t.Fatalf("first peer should be complete:\n%s", plain)
	}
	if !strings.Contains(plain, "… +") {
		t.Fatalf("expected overflow ellipsis:\n%s", plain)
	}
	if strings.Contains(plain, "192.0.2.20:443") {
		t.Fatalf("last peer should be omitted when capped:\n%s", plain)
	}
}

func TestProcPanel_KillConfirmSearch(t *testing.T) {
	p := NewProcPanel()
	var gotPID int32
	var gotForce bool
	p.killFn = func(pid int32, force bool) error {
		gotPID = pid
		gotForce = force
		return nil
	}
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 42, Name: "sshd"},
		{PID: 99, Name: "bash"},
	}})
	p.HandleKey("/")
	p.HandleKey("s")
	if !p.HandleKey("x") {
		t.Fatal("x should enter kill confirm")
	}
	if p.mode != procModeConfirmTERM {
		t.Fatalf("mode=%d", p.mode)
	}
	plain := stripANSI(p.View(80, 12))
	if !strings.Contains(plain, "SIGTERM") || !strings.Contains(plain, "42") {
		t.Fatalf("confirm view:\n%s", plain)
	}
	if !p.HandleKey("n") {
		t.Fatal("n cancels")
	}
	if p.mode != procModeSearch {
		t.Fatalf("mode after cancel=%d", p.mode)
	}

	p.HandleKey("x")
	p.HandleKey("y")
	if gotPID != 42 || gotForce {
		t.Fatalf("kill pid=%d force=%v", gotPID, gotForce)
	}
	if p.mode != procModeSearch {
		t.Fatalf("mode after kill=%d", p.mode)
	}
	if !strings.Contains(p.statusMsg, "SIGTERM") {
		t.Fatalf("status=%q", p.statusMsg)
	}
}

func TestProcPanel_KillConfirmDetail(t *testing.T) {
	p := NewProcPanel()
	p.killFn = func(pid int32, force bool) error { return nil }
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "bash"},
	}})
	p.HandleKey("/")
	p.HandleKey("enter")
	if p.mode != procModeDetail {
		t.Fatalf("mode=%d", p.mode)
	}
	p.HandleKey("x")
	if p.mode != procModeConfirmTERM || p.killPID != 10 {
		t.Fatalf("mode=%d pid=%d", p.mode, p.killPID)
	}
	p.HandleKey("esc")
	if p.mode != procModeDetail {
		t.Fatalf("mode=%d", p.mode)
	}
}

func TestProcPanel_KillForcePrompt(t *testing.T) {
	p := NewProcPanel()
	calls := 0
	p.killFn = func(pid int32, force bool) error {
		calls++
		if !force {
			return fmt.Errorf("permission denied")
		}
		return nil
	}
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 7, Name: "sticky"},
	}})
	p.HandleKey("/")
	p.HandleKey("x")
	p.HandleKey("y")
	if p.mode != procModeConfirmKILL {
		t.Fatalf("mode=%d want confirm KILL", p.mode)
	}
	plain := stripANSI(p.View(80, 12))
	if !strings.Contains(plain, "permission denied") || !strings.Contains(plain, "Force SIGKILL") {
		t.Fatalf("force prompt:\n%s", plain)
	}
	p.HandleKey("y")
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	if p.mode != procModeSearch {
		t.Fatalf("mode=%d", p.mode)
	}
	if !strings.Contains(p.statusMsg, "SIGKILL") {
		t.Fatalf("status=%q", p.statusMsg)
	}
}

func TestProcPanel_SearchPreservesOrder(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 1, Name: "high", CPUPercent: 90},
		{PID: 2, Name: "mid", CPUPercent: 50},
		{PID: 3, Name: "low", CPUPercent: 10},
	}})
	p.HandleKey("/")
	if len(p.order) != 3 || p.order[0] != 1 || p.order[1] != 2 || p.order[2] != 3 {
		t.Fatalf("order=%v", p.order)
	}

	// Resample with reversed CPU ranking; search mode must keep PID order.
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 3, Name: "low", CPUPercent: 99},
		{PID: 2, Name: "mid", CPUPercent: 80},
		{PID: 1, Name: "high", CPUPercent: 1},
	}})
	if len(p.all) != 3 || p.all[0].PID != 1 || p.all[1].PID != 2 || p.all[2].PID != 3 {
		t.Fatalf("all order=%v", procPIDs(p.all))
	}
	if p.all[0].CPUPercent != 1 || p.all[2].CPUPercent != 99 {
		t.Fatalf("values not refreshed: %#v", p.all)
	}
	if len(p.filtered) != 3 || p.filtered[0].PID != 1 {
		t.Fatalf("filtered=%v", procPIDs(p.filtered))
	}

	p.HandleKey("esc")
	if p.mode != procModeList || p.order != nil {
		t.Fatalf("mode=%d order=%v", p.mode, p.order)
	}
}

func TestProcPanel_DetailRefreshesStats(t *testing.T) {
	p := NewProcPanel()
	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "app", CPUPercent: 1, MemPercent: 2, Cmdline: "/bin/app"},
		{PID: 1, PPID: 0, Name: "init"},
	}})
	p.HandleKey("/")
	p.HandleKey("a")
	p.HandleKey("enter")
	if p.mode != procModeDetail || p.focusPID != 10 {
		t.Fatalf("mode=%d focus=%d", p.mode, p.focusPID)
	}
	keep := p.SampleKeepPIDs()
	if !containsPID(keep, 10) {
		t.Fatalf("KeepPIDs=%v", keep)
	}

	p.Update(hoststats.Snapshot{Processes: []hoststats.ProcStat{
		{PID: 10, Name: "app", CPUPercent: 55.5, MemPercent: 33.3, Cmdline: "/bin/app", DiskBytesPerSec: 4096},
		{PID: 1, PPID: 0, Name: "init"},
	}})
	plain := stripANSI(p.View(80, 20))
	if !strings.Contains(plain, "55.5%") || !strings.Contains(plain, "33.3%") {
		t.Fatalf("detail did not refresh cpu/mem:\n%s", plain)
	}
	if !strings.Contains(plain, "4.0KB") {
		t.Fatalf("detail did not refresh disk:\n%s", plain)
	}
}

func containsPID(pids []int32, want int32) bool {
	for _, p := range pids {
		if p == want {
			return true
		}
	}
	return false
}

func TestRenderSparkCurve(t *testing.T) {
	s := renderSparkCurve([]float64{0, 50, 100}, 8, 3, 100)
	lines := strings.Split(s, "\n")
	if len(lines) != 3 {
		t.Fatalf("lines=%d want 3: %q", len(lines), s)
	}
	hasBraille := false
	wantW := 8 * sparkCellWidth()
	for _, line := range lines {
		if lipgloss.Width(line) != wantW {
			t.Fatalf("width=%d want %d: %q", lipgloss.Width(line), wantW, line)
		}
		for _, r := range line {
			if r == '▁' || r == '█' {
				t.Fatalf("bar glyph in curve output: %q", s)
			}
			if r >= 0x2800 && r <= 0x28FF {
				hasBraille = true
			}
		}
	}
	if !hasBraille {
		t.Fatalf("expected Braille glyphs: %q", s)
	}
}

func TestRenderSparkCurve_Empty(t *testing.T) {
	s := renderSparkCurve(nil, 5, 2, 100)
	lines := strings.Split(s, "\n")
	if len(lines) != 2 {
		t.Fatalf("lines=%d: %q", len(lines), s)
	}
	want := strings.Repeat(" ", 5*sparkCellWidth())
	for _, line := range lines {
		if line != want {
			t.Fatalf("want blank row: %q", line)
		}
	}
}

func TestRenderSparkChart_Axes(t *testing.T) {
	t0 := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 8, 16, 12, 0, 5, 0, time.UTC)
	t2 := time.Date(2026, 8, 16, 12, 0, 10, 0, time.UTC)
	s := renderSparkChart(
		[]float64{10, 50, 90},
		[]time.Time{t0, t1, t2},
		40, 5,
		percentSparkScale(),
	)
	for _, want := range []string{"100%", "50%", "0%", "12:00:00", "12:00:10"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
	lines := strings.Split(s, "\n")
	if len(lines) != 5 {
		t.Fatalf("lines=%d want 5:\n%s", len(lines), s)
	}
	if strings.HasPrefix(lines[0], " ") && !strings.Contains(lines[0], "100%") {
		t.Fatalf("Y label not on first curve row: %q", lines[0])
	}
	for i, line := range lines {
		if lipgloss.Width(line) != 40 {
			t.Fatalf("row %d width=%d want 40: %q", i, lipgloss.Width(line), line)
		}
	}
}

func TestFormatXAxis_HHMMSS(t *testing.T) {
	t0 := time.Date(2026, 8, 16, 13, 1, 3, 0, time.UTC)
	t1 := time.Date(2026, 8, 16, 13, 1, 8, 0, time.UTC)
	t2 := time.Date(2026, 8, 16, 13, 1, 13, 0, time.UTC)
	got := formatXAxis([]time.Time{t0, t1, t2}, 0, 36)
	if !strings.Contains(got, "13:01:03") || !strings.Contains(got, "13:01:13") {
		t.Fatalf("x-axis=%q", got)
	}
	if utf8.RuneCountInString(got) != 36 {
		t.Fatalf("width=%d: %q", utf8.RuneCountInString(got), got)
	}
}

func TestSparklinePanel_CPUView(t *testing.T) {
	p := NewCPUPanel()
	p.Update(hoststats.Snapshot{
		Time: time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC),
		CPU:  hoststats.CPUStat{Percent: 17.5},
	})
	view := p.View(48, 12)
	plain := stripANSI(view)
	if !strings.Contains(plain, "CPU   17.5%") {
		t.Fatalf("missing centered headline in:\n%s", plain)
	}
	if !strings.Contains(view, ansiCPU) {
		t.Fatalf("CPU value/curve should use CPU series color:\n%s", view)
	}
	for _, want := range []string{"100%", "0%", "12:00:00"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("missing %q in:\n%s", want, plain)
		}
	}
}

func TestFormatPanelHeadline(t *testing.T) {
	got := formatPanelHeadline("CPU", "15.0%")
	if got != "CPU   15.0%" {
		t.Fatalf("got %q", got)
	}
}

func TestCenterPanelLine(t *testing.T) {
	got := centerPanelLine("CPU   15.0%", 20)
	if utf8.RuneCountInString(got) != 20 {
		t.Fatalf("width=%d: %q", utf8.RuneCountInString(got), got)
	}
	if !strings.Contains(got, "CPU   15.0%") {
		t.Fatalf("%q", got)
	}
	if strings.TrimSpace(got) != "CPU   15.0%" {
		t.Fatalf("unexpected trim: %q", got)
	}
	left := strings.Index(got, "CPU")
	right := 20 - left - utf8.RuneCountInString("CPU   15.0%")
	if left < 1 || right < 1 {
		t.Fatalf("not centered left=%d right=%d: %q", left, right, got)
	}
}

func TestMemDiskNet_Headlines(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	mem := NewMemPanel()
	mem.Update(hoststats.Snapshot{
		Time: now,
		Mem:  hoststats.MemStat{UsedPercent: 62.6, Used: 10 << 30, Total: 16 << 30},
	})
	if !strings.Contains(stripANSI(mem.View(48, 12)), "MEM   62.6%") {
		t.Fatalf("MEM headline:\n%s", mem.View(48, 12))
	}
	if !strings.Contains(mem.View(48, 12), ansiMEM) {
		t.Fatalf("MEM value/curve should use MEM series color")
	}

	disk := NewDiskPanel()
	disk.Update(hoststats.Snapshot{
		Time: now,
		Disk: hoststats.DiskStat{
			Mount: "/", UsedPercent: 83.3,
			ReadBytesPerSec: 1 << 20, WriteBytesPerSec: 128 << 10,
		},
	})
	diskPlain := stripANSI(disk.View(48, 12))
	if !strings.Contains(diskPlain, "DISK   83.3%") {
		t.Fatalf("DISK headline:\n%s", diskPlain)
	}
	if strings.Contains(diskPlain, "100%") {
		t.Fatalf("DISK chart should use rate Y-axis, not percent:\n%s", diskPlain)
	}
	if !strings.Contains(diskPlain, "/s") {
		t.Fatalf("DISK chart missing rate ticks:\n%s", diskPlain)
	}
	if !strings.Contains(diskPlain, "r ") || !strings.Contains(diskPlain, "w ") {
		t.Fatalf("DISK r/w rates:\n%s", diskPlain)
	}

	net := NewNetPanel()
	net.Update(hoststats.Snapshot{
		Time: now,
		Net: hoststats.NetStat{
			Name: "all", RxBytesPerSec: 3400, TxBytesPerSec: 3700,
			ErrIn: 0, ErrOut: 13, DropIn: 0, DropOut: 338,
		},
	})
	plain := stripANSI(net.View(56, 14))
	if !strings.Contains(plain, "NET   ↓") || !strings.Contains(plain, "↑") {
		t.Fatalf("NET headline:\n%s", plain)
	}
	if !strings.Contains(plain, "errs 0/13") || !strings.Contains(plain, "drops 0/338") {
		t.Fatalf("NET errs/drops:\n%s", plain)
	}
	netView := net.View(56, 14)
	if !strings.Contains(netView, ansiIn) || !strings.Contains(netView, ansiOut) {
		t.Fatalf("NET ↓/↑ should match dual-series colors:\n%s", netView)
	}
}

func TestDiskPanel_ReadWriteCurves(t *testing.T) {
	now := time.Date(2026, 8, 16, 15, 0, 0, 0, time.UTC)
	p := NewDiskPanel()
	for i := 0; i < 5; i++ {
		p.Update(hoststats.Snapshot{
			Time: now.Add(time.Duration(i) * time.Second),
			Disk: hoststats.DiskStat{
				Mount: "/", UsedPercent: 87.9,
				ReadBytesPerSec: float64(i+1) * 500000, WriteBytesPerSec: float64(i+1) * 1000,
			},
		})
	}
	if len(p.histR) != 5 || len(p.histW) != 5 {
		t.Fatalf("histR=%d histW=%d", len(p.histR), len(p.histW))
	}
	view := p.View(64, 14)
	if !strings.Contains(view, "\x1b[") {
		t.Fatalf("expected ANSI colors for read/write curves:\n%s", view)
	}
	plain := stripANSI(view)
	if !strings.Contains(plain, "DISK   87.9%") {
		t.Fatalf("usage title:\n%s", plain)
	}
	if strings.Contains(plain, "100%") {
		t.Fatalf("percent Y-axis still present:\n%s", plain)
	}
	// Independent Y scales: left tracks high read (~2.5MB/s), right tracks low write (~5KB/s).
	if !strings.Contains(plain, "MB/s") || !strings.Contains(plain, "KB/s") && !strings.Contains(plain, "B/s") {
		t.Fatalf("expected dual rate tick units:\n%s", plain)
	}
}

func TestNetPanel_RxTxCurves(t *testing.T) {
	now := time.Date(2026, 8, 16, 15, 0, 0, 0, time.UTC)
	p := NewNetPanel()
	for i := 0; i < 5; i++ {
		p.Update(hoststats.Snapshot{
			Time: now.Add(time.Duration(i) * time.Second),
			Net: hoststats.NetStat{
				Name:          "all",
				RxBytesPerSec: float64(i+1) * 500000, TxBytesPerSec: float64(i+1) * 1000,
			},
		})
	}
	if len(p.histRx) != 5 || len(p.histTx) != 5 {
		t.Fatalf("histRx=%d histTx=%d", len(p.histRx), len(p.histTx))
	}
	view := p.View(64, 14)
	if !strings.Contains(view, ansiIn) || !strings.Contains(view, ansiOut) {
		t.Fatalf("expected in/out colors matching DISK series:\n%s", view)
	}
	plain := stripANSI(view)
	if !strings.Contains(plain, "↓") || !strings.Contains(plain, "↑") {
		t.Fatalf("headline arrows:\n%s", plain)
	}
	if strings.Contains(plain, "100%") {
		t.Fatalf("percent Y-axis still present:\n%s", plain)
	}
	if !strings.Contains(plain, "MB/s") {
		t.Fatalf("left axis should scale to rx (~2.5MB/s):\n%s", plain)
	}
}

func TestRenderSparkChartDual_IndependentAxes(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	times := []time.Time{now, now.Add(time.Second), now.Add(2 * time.Second)}
	histL := []float64{1000000, 2000000, 1500000} // ~MB/s scale
	histR := []float64{1000, 2000, 1500}          // ~KB/s scale
	chart := renderSparkChartDual(histL, histR, times, 48, 8, rateSparkScale())
	if !strings.Contains(chart, "\x1b[") {
		t.Fatalf("missing colors:\n%s", chart)
	}
	plain := stripANSI(chart)
	if !strings.Contains(plain, "MB/s") {
		t.Fatalf("left axis should show MB/s for read:\n%s", plain)
	}
	if !strings.Contains(plain, "KB/s") && !strings.Contains(plain, "1.5KB/s") && !strings.Contains(plain, "2.0KB/s") {
		// write max ~2KB/s
		if !strings.Contains(plain, "B/s") {
			t.Fatalf("right axis missing rate ticks:\n%s", plain)
		}
	}
}

func TestRenderSparkChartDual_AxisMatchesRWColors(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	times := []time.Time{now, now.Add(time.Second)}
	chart := renderSparkChartDual(nil, nil, times, 48, 6, rateSparkScale())
	lines := strings.Split(chart, "\n")
	if len(lines) < 2 {
		t.Fatalf("short chart:\n%s", chart)
	}

	top := lines[0]
	if !strings.Contains(top, ansiIn) {
		t.Fatalf("left Y ticks should use read (r) color:\n%q", top)
	}
	if !strings.Contains(top, ansiOut) {
		t.Fatalf("right Y ticks should use write (w) color:\n%q", top)
	}

	x := lines[len(lines)-1]
	if strings.Contains(x, ansiIn) || strings.Contains(x, ansiOut) {
		t.Fatalf("X-axis should stay faint, not r/w colors:\n%q", x)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
