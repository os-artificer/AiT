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
	for _, id := range []string{IDCPU, IDMem, IDDisk, IDNet, IDProc} {
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
	for _, want := range []string{IDCPU, IDMem, IDDisk, IDNet, IDProc} {
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
	row := FormatProcRow(ProcStat{PID: 1, Name: "init", CPUPercent: 1.5, MemPercent: 2.5})
	if !strings.Contains(row, "init") || !strings.Contains(row, "1.5") {
		t.Fatalf("%q", row)
	}
}
