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
	"strings"
	"testing"
)

func TestProcessQuery_DefaultSortLimit(t *testing.T) {
	procs := []ProcStat{
		{PID: 1, Name: "a", CPUPercent: 1, MemPercent: 50, NetBytesPerSec: 100, DiskBytesPerSec: 1},
		{PID: 2, Name: "b", CPUPercent: 90, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 3, Name: "c", CPUPercent: 90, MemPercent: 40, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 4, Name: "d", CPUPercent: 5, MemPercent: 5, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 5, Name: "e", CPUPercent: 4, MemPercent: 4, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 6, Name: "f", CPUPercent: 3, MemPercent: 3, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 7, Name: "g", CPUPercent: 2, MemPercent: 2, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 8, Name: "h", CPUPercent: 1.5, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 9, Name: "i", CPUPercent: 1.2, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 10, Name: "j", CPUPercent: 1.1, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 11, Name: "k", CPUPercent: 0.5, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
		{PID: 12, Name: "l", CPUPercent: 0.1, MemPercent: 1, NetBytesPerSec: 0, DiskBytesPerSec: 0},
	}
	got := SortAndLimitProcesses(procs, ProcessQuery{})
	if len(got) != DefaultProcessLimit {
		t.Fatalf("len=%d want %d", len(got), DefaultProcessLimit)
	}
	// CPU 90 with higher mem first, then other CPU 90.
	if got[0].PID != 3 || got[1].PID != 2 {
		t.Fatalf("order pid=%d,%d want 3,2", got[0].PID, got[1].PID)
	}
	// 12th process truncated.
	for _, p := range got {
		if p.PID == 12 {
			t.Fatal("pid 12 should be truncated")
		}
	}
}

func TestProcessQuery_SortMemLimit(t *testing.T) {
	procs := []ProcStat{
		{PID: 1, CPUPercent: 99, MemPercent: 10},
		{PID: 2, CPUPercent: 1, MemPercent: 80},
		{PID: 3, CPUPercent: 50, MemPercent: 50},
	}
	got := SortAndLimitProcesses(procs, ProcessQuery{Sort: SortMem, Limit: 2})
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].PID != 2 || got[1].PID != 3 {
		t.Fatalf("got %d,%d", got[0].PID, got[1].PID)
	}
}

func TestProcessQuery_KeepPIDsOutsideTopN(t *testing.T) {
	procs := []ProcStat{
		{PID: 1, Name: "hot", CPUPercent: 90},
		{PID: 2, Name: "warm", CPUPercent: 50},
		{PID: 99, Name: "cold", CPUPercent: 0.1, MemPercent: 12},
	}
	got := SortAndLimitProcesses(procs, ProcessQuery{Limit: 2, KeepPIDs: []int32{99}})
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].PID != 99 {
		t.Fatalf("keep should be first: %#v", got)
	}
	if got[0].MemPercent != 12 {
		t.Fatalf("keep stats not fresh: %#v", got[0])
	}
	if got[1].PID != 1 {
		t.Fatalf("fill from top sort: %#v", got)
	}
}

func TestProcessQuery_AllReturnsEveryProcess(t *testing.T) {
	procs := make([]ProcStat, 25)
	for i := range procs {
		procs[i] = ProcStat{PID: int32(i + 1), CPUPercent: float64(25 - i)}
	}
	got := SortAndLimitProcesses(procs, ProcessQuery{All: true, Limit: 3})
	if len(got) != 25 {
		t.Fatalf("len=%d want 25 (All ignores Limit)", len(got))
	}
	if got[0].PID != 1 || got[24].PID != 25 {
		t.Fatalf("sorted ends: first=%d last=%d", got[0].PID, got[24].PID)
	}
	q := NormalizeProcessQuery(ProcessQuery{All: true, Limit: 99})
	if q.Limit != 0 {
		t.Fatalf("All should clear Limit: got %d", q.Limit)
	}
}

func TestParseSortKey(t *testing.T) {
	k, err := ParseSortKey("MEM")
	if err != nil || k != SortMem {
		t.Fatalf("k=%q err=%v", k, err)
	}
	if _, err := ParseSortKey("gpu"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatBytes(t *testing.T) {
	if FormatBytes(500) != "500B" {
		t.Fatalf("%s", FormatBytes(500))
	}
	if FormatBytes(2048) != "2.0KB" {
		t.Fatalf("%s", FormatBytes(2048))
	}
}

func TestSectionByID(t *testing.T) {
	for _, id := range []string{IDCPU, IDMem, IDDisk, IDNet, IDProto, IDProc} {
		sec := SectionByID(id)
		if sec == nil || sec.ID() != id {
			t.Fatalf("SectionByID(%q)=%v", id, sec)
		}
	}
	if SectionByID("nope") != nil {
		t.Fatal("expected nil")
	}
}

func TestListSources(t *testing.T) {
	srcs := ListSources()
	ids := map[string]bool{}
	for _, s := range srcs {
		ids[s.ID] = true
	}
	for _, want := range []string{IDCPU, IDMem, IDDisk, IDNet, IDProto, IDProc} {
		if !ids[want] {
			t.Fatalf("missing source %s", want)
		}
	}
}

func TestTruncateName(t *testing.T) {
	if got := TruncateName("short", 20); got != "short" {
		t.Fatalf("%q", got)
	}
	if got := TruncateName("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("%q", got)
	}
	if got := TruncateName("中文进程名称超长测试", 6); got != "中文进..." {
		t.Fatalf("%q", got)
	}
	if got := TruncateName("ab", 2); got != "ab" {
		t.Fatalf("%q", got)
	}
	if got := TruncateName("abcd", 3); got != "abc" {
		t.Fatalf("%q", got)
	}
	if got := TruncateName("x", 0); got != "" {
		t.Fatalf("%q", got)
	}
}

func TestFormatProcRow(t *testing.T) {
	row := FormatProcRow(ProcStat{
		PID: 1, Name: "init", CPUPercent: 1.5, MemPercent: 2.5,
		Listen: []ProcListen{{Port: 80, Conns: 3}},
	})
	if !strings.Contains(row, "init") || !strings.Contains(row, "1.5") {
		t.Fatalf("%q", row)
	}
	if !strings.Contains(row, ":80×3") {
		t.Fatalf("missing peer: %q", row)
	}
	if strings.Contains(row, "NET/s") {
		t.Fatalf("unexpected NET/s in row: %q", row)
	}
	if !strings.Contains(ProcTableHeader, "PEER") || !strings.Contains(ProcTableHeader, "LOG") {
		t.Fatalf("header=%q", ProcTableHeader)
	}
	if strings.Contains(ProcTableHeader, "NET/s") {
		t.Fatalf("header=%q", ProcTableHeader)
	}
}

func TestFormatProcPeer_Server(t *testing.T) {
	got := FormatProcPeer(ProcStat{
		Listen:  []ProcListen{{Port: 8080, Conns: 12}, {Port: 443, Conns: 0}},
		Remotes: []string{"1.2.3.4:99"}, // ignored when Listen present
	})
	if got != ":8080×12,:443×0" {
		t.Fatalf("%q", got)
	}
}

func TestFormatProcPeer_Client(t *testing.T) {
	got := FormatProcPeer(ProcStat{Remotes: []string{"10.0.0.1:443", "10.0.0.2:80"}})
	if !strings.Contains(got, "10.0.0.1:443") || !strings.Contains(got, "10.0.0.2:80") {
		t.Fatalf("%q", got)
	}
}

func TestFormatProcPeer_Empty(t *testing.T) {
	if got := FormatProcPeer(ProcStat{}); got != "" {
		t.Fatalf("%q", got)
	}
}

func TestFormatProcPeer_Truncates(t *testing.T) {
	long := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		long = append(long, "192.168.100.200:65535")
	}
	got := FormatProcPeer(ProcStat{Remotes: long})
	runes := []rune(got)
	if len(runes) > ProcPeerColWidth {
		t.Fatalf("width=%d > %d: %q", len(runes), ProcPeerColWidth, got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncation: %q", got)
	}
}

func TestFormatProcPeerLines_OnePerRecord(t *testing.T) {
	got := FormatProcPeerLines(ProcStat{
		Listen:  []ProcListen{{Port: 8080, Conns: 12}, {Port: 443, Conns: 0}},
		Remotes: []string{"1.2.3.4:99"},
	})
	if len(got) != 2 || got[0] != ":8080×12" || got[1] != ":443×0" {
		t.Fatalf("%q", got)
	}

	remotes := FormatProcPeerLines(ProcStat{Remotes: []string{"10.0.0.1:443", "10.0.0.2:80"}})
	if len(remotes) != 2 || remotes[0] != "10.0.0.1:443" || remotes[1] != "10.0.0.2:80" {
		t.Fatalf("%q", remotes)
	}
	if FormatProcPeerLines(ProcStat{}) != nil {
		t.Fatal("empty should be nil")
	}
}

func TestAggregateProcPeers_Server(t *testing.T) {
	got := aggregateProcPeers([]procConn{
		{Pid: 1, Status: ConnStatusListen, LPort: 8080},
		{Pid: 1, Status: ConnStatusListen, LPort: 443},
		{Pid: 1, Status: ConnStatusEstablished, LPort: 8080, RIP: "1.1.1.1", RPort: 50000},
		{Pid: 1, Status: ConnStatusEstablished, LPort: 8080, RIP: "2.2.2.2", RPort: 50001},
	})
	p, ok := got[1]
	if !ok {
		t.Fatal("missing pid 1")
	}
	if len(p.Listen) != 2 || len(p.Remotes) != 0 {
		t.Fatalf("%+v", p)
	}
	if p.Listen[0].Port != 443 || p.Listen[0].Conns != 0 {
		t.Fatalf("port 443: %+v", p.Listen[0])
	}
	if p.Listen[1].Port != 8080 || p.Listen[1].Conns != 2 {
		t.Fatalf("port 8080: %+v", p.Listen[1])
	}
}

func TestAggregateProcPeers_Client(t *testing.T) {
	got := aggregateProcPeers([]procConn{
		{Pid: 7, Status: ConnStatusEstablished, LPort: 54321, RIP: "8.8.8.8", RPort: 53},
		{Pid: 7, Status: ConnStatusEstablished, LPort: 54322, RIP: "8.8.8.8", RPort: 53}, // same remote
	})
	p, ok := got[7]
	if !ok {
		t.Fatal("missing pid 7")
	}
	if len(p.Listen) != 0 || len(p.Remotes) != 1 || p.Remotes[0] != "8.8.8.8:53" {
		t.Fatalf("%+v", p)
	}
}

func TestAggregateProcPeers_Empty(t *testing.T) {
	got := aggregateProcPeers([]procConn{
		{Pid: 3, Status: "TIME_WAIT", LPort: 1},
		{Pid: 0, Status: ConnStatusListen, LPort: 22},
	})
	if len(got) != 0 {
		t.Fatalf("%+v", got)
	}
}
