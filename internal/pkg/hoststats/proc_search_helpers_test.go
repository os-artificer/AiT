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

import "testing"

func TestMatchProc_EmptyQuery(t *testing.T) {
	p := ProcStat{PID: 1, Name: "init"}
	if !MatchProc("", p) || !MatchProc("  ", p) {
		t.Fatal("empty query should match")
	}
}

func TestMatchProc_SubstringName(t *testing.T) {
	p := ProcStat{PID: 42, Name: "sshd", Cmdline: "/usr/sbin/sshd -D"}
	if !MatchProc("ssh", p) {
		t.Fatal("expected name substring match")
	}
	if MatchProc("zzz", p) {
		t.Fatal("unexpected match")
	}
}

func TestMatchProc_SubstringCmdline(t *testing.T) {
	p := ProcStat{PID: 7, Name: "go", Cmdline: "/usr/local/bin/go test ./..."}
	if !MatchProc("test", p) {
		t.Fatal("expected cmdline substring match")
	}
}

func TestMatchProc_PID(t *testing.T) {
	p := ProcStat{PID: 1234, Name: "foo"}
	if !MatchProc("123", p) {
		t.Fatal("expected PID match")
	}
	if !MatchProc("1234", p) {
		t.Fatal("expected full PID match")
	}
}

func TestMatchProc_CaseInsensitive(t *testing.T) {
	p := ProcStat{PID: 1, Name: "Dockerd", Cmdline: "Docker Desktop"}
	if !MatchProc("docker", p) {
		t.Fatal("expected case-insensitive match")
	}
}

func TestMatchProc_Subsequence(t *testing.T) {
	p := ProcStat{PID: 9, Name: "chromium", Cmdline: ""}
	if !MatchProc("crm", p) {
		t.Fatal("expected subsequence match on name")
	}
}

func TestFilterProcs_PreservesOrder(t *testing.T) {
	in := []ProcStat{
		{PID: 1, Name: "a"},
		{PID: 2, Name: "bb"},
		{PID: 3, Name: "abc"},
	}
	out := FilterProcs("b", in)
	if len(out) != 2 || out[0].PID != 2 || out[1].PID != 3 {
		t.Fatalf("got %#v", out)
	}
}
