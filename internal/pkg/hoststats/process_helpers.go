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
	"fmt"
	"sort"
	"strings"
)

// NormalizeProcessQuery fills sort defaults and clamps Limit.
// When All is true, Limit is cleared to 0 and not capped.
func NormalizeProcessQuery(q ProcessQuery) ProcessQuery {
	if q.All {
		q.Limit = 0
	} else if q.Limit <= 0 {
		q.Limit = DefaultProcessLimit
	}
	if !q.All && q.Limit > MaxProcessLimit {
		q.Limit = MaxProcessLimit
	}
	q.Sort = strings.ToLower(strings.TrimSpace(q.Sort))
	switch q.Sort {
	case "", SortCPU, SortMem, SortNet, SortDisk:
		if q.Sort == "" {
			q.Sort = SortCPU
		}
	default:
		q.Sort = SortCPU
	}
	return q
}

// ParseSortKey validates a --sort value.
// Returns an error when s is non-empty and not a known key.
func ParseSortKey(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return SortCPU, nil
	}
	switch s {
	case SortCPU, SortMem, SortNet, SortDisk:
		return s, nil
	default:
		return "", fmt.Errorf("unknown sort key %q", s)
	}
}

// SortAndLimitProcesses sorts procs by q and returns at most q.Limit rows
// (or every process when q.All is true).
// When q.KeepPIDs is set, those PIDs are included first (still-running only),
// then remaining slots are filled from the sorted list.
func SortAndLimitProcesses(procs []ProcStat, q ProcessQuery) []ProcStat {
	q = NormalizeProcessQuery(q)
	if q.All {
		return sortKeepAll(procs, q)
	}
	if len(q.KeepPIDs) == 0 {
		return sortAndTruncate(procs, q)
	}
	return sortLimitKeep(procs, q)
}

func sortAndTruncate(procs []ProcStat, q ProcessQuery) []ProcStat {
	out := sortProcs(procs, q.Sort)
	if len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out
}

func sortProcs(procs []ProcStat, primary string) []ProcStat {
	out := append([]ProcStat(nil), procs...)
	sort.SliceStable(out, func(i, j int) bool {
		return lessProc(out[i], out[j], primary)
	})
	return out
}

func sortKeepAll(procs []ProcStat, q ProcessQuery) []ProcStat {
	return preferKeepPIDs(procs, q, 0)
}

func sortLimitKeep(procs []ProcStat, q ProcessQuery) []ProcStat {
	return preferKeepPIDs(procs, q, q.Limit)
}

// preferKeepPIDs puts still-running KeepPIDs first, then the sorted remainder.
// limit <= 0 means no cap (every process).
func preferKeepPIDs(procs []ProcStat, q ProcessQuery, limit int) []ProcStat {
	byPID := make(map[int32]ProcStat, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}

	capHint := len(procs)
	if limit > 0 {
		capHint = limit
	}
	out := make([]ProcStat, 0, capHint)
	seen := make(map[int32]bool, capHint)
	for _, pid := range q.KeepPIDs {
		if limit > 0 && len(out) >= limit {
			break
		}
		p, ok := byPID[pid]
		if !ok || seen[pid] {
			continue
		}
		out = append(out, p)
		seen[pid] = true
	}

	for _, p := range sortProcs(procs, q.Sort) {
		if limit > 0 && len(out) >= limit {
			break
		}
		if seen[p.PID] {
			continue
		}
		out = append(out, p)
		seen[p.PID] = true
	}
	return out
}

func lessProc(a, b ProcStat, primary string) bool {
	keys := sortKeyOrder(primary)
	for _, k := range keys {
		av, bv := procMetric(a, k), procMetric(b, k)
		if av != bv {
			return av > bv // descending
		}
	}
	return a.PID < b.PID
}

func sortKeyOrder(primary string) []string {
	all := []string{SortCPU, SortMem, SortNet, SortDisk}
	out := make([]string, 0, len(all))
	out = append(out, primary)
	for _, k := range all {
		if k != primary {
			out = append(out, k)
		}
	}
	return out
}

func procMetric(p ProcStat, key string) float64 {
	switch key {
	case SortMem:
		return p.MemPercent
	case SortNet:
		return p.NetBytesPerSec
	case SortDisk:
		return p.DiskBytesPerSec
	default:
		return p.CPUPercent
	}
}
