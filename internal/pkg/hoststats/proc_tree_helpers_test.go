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

func TestProcAncestors_ChainRootToSelf(t *testing.T) {
	all := []ProcStat{
		{PID: 1, PPID: 0, Name: "init"},
		{PID: 10, PPID: 1, Name: "bash"},
		{PID: 20, PPID: 10, Name: "vim"},
	}
	got := ProcAncestors(all, 20)
	if len(got) != 3 {
		t.Fatalf("len=%d want 3", len(got))
	}
	if got[0].PID != 1 || got[1].PID != 10 || got[2].PID != 20 {
		t.Fatalf("chain=%v", pids(got))
	}
}

func TestProcAncestors_MissingParentStops(t *testing.T) {
	all := []ProcStat{
		{PID: 20, PPID: 99, Name: "orphan"},
	}
	got := ProcAncestors(all, 20)
	if len(got) != 1 || got[0].PID != 20 {
		t.Fatalf("got %#v", got)
	}
}

func TestProcAncestors_UnknownPID(t *testing.T) {
	if ProcAncestors(nil, 1) != nil {
		t.Fatal("want nil")
	}
}

func TestProcChildren_SortedByCPU(t *testing.T) {
	all := []ProcStat{
		{PID: 1, PPID: 0, Name: "init"},
		{PID: 2, PPID: 1, Name: "low", CPUPercent: 1},
		{PID: 3, PPID: 1, Name: "high", CPUPercent: 90},
		{PID: 4, PPID: 2, Name: "other"},
	}
	kids := ProcChildren(all, 1)
	if len(kids) != 2 || kids[0].PID != 3 || kids[1].PID != 2 {
		t.Fatalf("kids=%v", pids(kids))
	}
}

func pids(ps []ProcStat) []int32 {
	out := make([]int32, len(ps))
	for i, p := range ps {
		out[i] = p.PID
	}
	return out
}
