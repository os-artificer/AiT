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

import "sort"

// ProcAncestors returns the ancestor chain from root toward self (inclusive),
// walking PPID within the provided sample. Missing parents stop the walk.
func ProcAncestors(all []ProcStat, pid int32) []ProcStat {
	byPID := make(map[int32]ProcStat, len(all))
	for _, p := range all {
		byPID[p.PID] = p
	}
	self, ok := byPID[pid]
	if !ok {
		return nil
	}

	chain := []ProcStat{self}
	seen := map[int32]bool{pid: true}
	cur := self.PPID
	for cur > 0 && !seen[cur] {
		parent, ok := byPID[cur]
		if !ok {
			break
		}
		seen[cur] = true
		chain = append(chain, parent)
		cur = parent.PPID
	}

	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// ProcChildren returns direct children of pid (PPID == pid), sorted by CPU descending.
func ProcChildren(all []ProcStat, pid int32) []ProcStat {
	var kids []ProcStat
	for _, p := range all {
		if p.PPID == pid {
			kids = append(kids, p)
		}
	}
	sort.SliceStable(kids, func(i, j int) bool {
		return kids[i].CPUPercent > kids[j].CPUPercent
	})
	return kids
}
