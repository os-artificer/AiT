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
	"strconv"
	"strings"
)

// MatchProc reports whether p matches query (case-insensitive substring or subsequence
// against Name, Cmdline, or decimal PID). Empty query matches all.
func MatchProc(query string, p ProcStat) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return true
	}
	qFold := foldASCII(q)
	if matchFuzzy(qFold, foldASCII(p.Name)) {
		return true
	}
	if matchFuzzy(qFold, foldASCII(p.Cmdline)) {
		return true
	}
	return matchFuzzy(qFold, strconv.FormatInt(int64(p.PID), 10))
}

// FilterProcs returns processes matching query, preserving input order.
func FilterProcs(query string, procs []ProcStat) []ProcStat {
	out := make([]ProcStat, 0, len(procs))
	for _, p := range procs {
		if MatchProc(query, p) {
			out = append(out, p)
		}
	}
	return out
}

func matchFuzzy(query, hay string) bool {
	if query == "" {
		return true
	}
	if strings.Contains(hay, query) {
		return true
	}
	return subsequenceMatch(query, hay)
}

func subsequenceMatch(query, hay string) bool {
	qRunes := []rune(query)
	qi := 0
	for _, r := range hay {
		if qi >= len(qRunes) {
			break
		}
		if r == qRunes[qi] {
			qi++
		}
	}
	return qi == len(qRunes)
}

func foldASCII(s string) string {
	return strings.ToLower(s)
}
