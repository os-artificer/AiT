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
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

func TestAttachProcLogs_PIDMatch(t *testing.T) {
	procs := []ProcStat{{PID: 42, Name: "sshd"}, {PID: 99, Name: "bash"}}
	recs := []logsrc.Record{
		{PID: 42, Message: "old", Time: time.Unix(1, 0)},
		{PID: 42, Message: "newest sshd", Time: time.Unix(2, 0)},
		{PID: 7, Message: "other", Time: time.Unix(3, 0)},
	}
	attachProcLogs(procs, recs)
	if procs[0].LogLine != "newest sshd" {
		t.Fatalf("logline=%q", procs[0].LogLine)
	}
	if len(procs[0].Logs) != 2 || procs[0].Logs[1].Message != "newest sshd" {
		t.Fatalf("logs=%+v", procs[0].Logs)
	}
	if procs[1].LogLine != "" {
		t.Fatalf("bash should have no logs: %q", procs[1].LogLine)
	}
}

func TestAttachProcLogs_NameFallback(t *testing.T) {
	procs := []ProcStat{{PID: 1, Name: "sshd"}}
	recs := []logsrc.Record{
		{PID: 0, Process: "sshd", Message: "by name"},
	}
	attachProcLogs(procs, recs)
	if procs[0].LogLine != "by name" {
		t.Fatalf("logline=%q", procs[0].LogLine)
	}
}

func TestAttachProcLogs_PIDBeatsName(t *testing.T) {
	procs := []ProcStat{{PID: 42, Name: "sshd"}}
	recs := []logsrc.Record{
		{PID: 99, Process: "sshd", Message: "wrong process"},
		{PID: 42, Process: "sshd", Message: "this pid"},
	}
	attachProcLogs(procs, recs)
	if procs[0].LogLine != "this pid" {
		t.Fatalf("logline=%q", procs[0].LogLine)
	}
}

func TestAttachProcLogs_DetailCap(t *testing.T) {
	procs := []ProcStat{{PID: 5, Name: "x"}}
	recs := make([]logsrc.Record, ProcLogDetailMax+5)
	for i := range recs {
		recs[i] = logsrc.Record{PID: 5, Message: strings.Repeat("m", 1) + string(rune('a'+i%26))}
	}
	attachProcLogs(procs, recs)
	if len(procs[0].Logs) != ProcLogDetailMax {
		t.Fatalf("len=%d want %d", len(procs[0].Logs), ProcLogDetailMax)
	}
}

func TestFormatProcRow_IncludesLog(t *testing.T) {
	row := FormatProcRow(ProcStat{
		PID: 1, Name: "init", CPUPercent: 1.5, MemPercent: 2.5,
		Listen:  []ProcListen{{Port: 80, Conns: 3}},
		LogLine: "accepted connection from 10.0.0.1",
	})
	if !strings.Contains(row, "init") || !strings.Contains(row, ":80×3") {
		t.Fatalf("%q", row)
	}
	if !strings.Contains(row, "accepted") {
		t.Fatalf("missing log: %q", row)
	}
	hdr := FormatProcTableHeader(ProcLogColWidth)
	if !strings.Contains(hdr, "PEER") || !strings.Contains(hdr, "LOG") {
		t.Fatalf("header=%q", hdr)
	}
	if strings.Contains(hdr, "NET/s") {
		t.Fatalf("header=%q", hdr)
	}
}

func TestFlattenLogMessage_Newlines(t *testing.T) {
	p := ProcStat{LogLine: "a\nb\r\nc"}
	got := FormatProcLog(p, 20)
	if strings.Contains(got, "\n") || !strings.Contains(got, "a b c") {
		t.Fatalf("%q", got)
	}
}

func TestProcRowMetaWidth(t *testing.T) {
	const logW = 10
	row := FormatProcRowWidth(ProcStat{PID: 1, Name: "n", LogLine: "x"}, logW)
	if n := len(row); n != ProcRowMetaWidth+logW {
		t.Fatalf("row len=%d want %d %q", n, ProcRowMetaWidth+logW, row)
	}
	hdr := FormatProcTableHeader(logW)
	if n := len(hdr); n != ProcRowMetaWidth+logW {
		t.Fatalf("hdr len=%d want %d %q", n, ProcRowMetaWidth+logW, hdr)
	}
}
