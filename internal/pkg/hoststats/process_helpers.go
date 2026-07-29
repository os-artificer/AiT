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

// NormalizeProcessQuery fills defaults and clamps Limit.
func NormalizeProcessQuery(q ProcessQuery) ProcessQuery {
	if q.Limit <= 0 {
		q.Limit = DefaultProcessLimit
	}
	if q.Limit > MaxProcessLimit {
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

// SortAndLimitProcesses sorts procs by q and returns at most q.Limit rows.
func SortAndLimitProcesses(procs []ProcStat, q ProcessQuery) []ProcStat {
	q = NormalizeProcessQuery(q)
	out := append([]ProcStat(nil), procs...)
	sort.SliceStable(out, func(i, j int) bool {
		return lessProc(out[i], out[j], q.Sort)
	})
	if len(out) > q.Limit {
		out = out[:q.Limit]
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
