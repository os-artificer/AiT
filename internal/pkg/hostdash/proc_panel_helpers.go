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
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

const (
	procLogColMinWidth = 8
	procLogStampEmpty  = "--:--:--"
	procDetailNone     = "  (none)"
	procDetailMoreFmt  = "  … +%d more"
	procDetailItemFmt  = "  %s\n"
	procLogStampSep    = " "
)

func isProcSearchRune(key string) bool {
	if key == "" || strings.HasPrefix(key, "ctrl+") || strings.HasPrefix(key, "alt+") {
		return false
	}
	if utf8.RuneCountInString(key) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(key)
	return unicode.IsPrint(r)
}

func procPIDs(procs []hoststats.ProcStat) []int32 {
	out := make([]int32, len(procs))
	for i, p := range procs {
		out[i] = p.PID
	}
	return out
}

func appendKeepPID(pids []int32, pid int32) []int32 {
	for _, existing := range pids {
		if existing == pid {
			return pids
		}
	}
	return append(pids, pid)
}

// mergeProcsPreserveOrder keeps prior PID order, refreshes stats from fresh,
// drops gone PIDs, and appends newly seen PIDs in fresh order.
func mergeProcsPreserveOrder(order []int32, fresh []hoststats.ProcStat) []hoststats.ProcStat {
	byPID := make(map[int32]hoststats.ProcStat, len(fresh))
	for _, p := range fresh {
		byPID[p.PID] = p
	}

	out := make([]hoststats.ProcStat, 0, len(fresh))
	seen := make(map[int32]bool, len(fresh))
	for _, pid := range order {
		p, ok := byPID[pid]
		if !ok || seen[pid] {
			continue
		}
		out = append(out, p)
		seen[pid] = true
	}
	for _, p := range fresh {
		if seen[p.PID] {
			continue
		}
		out = append(out, p)
		seen[p.PID] = true
	}
	return out
}

func procLogColumnWidth(contentW int) int {
	logW := sparkInnerWidth(contentW) - hoststats.ProcRowMetaWidth
	if logW < procLogColMinWidth {
		return procLogColMinWidth
	}
	return logW
}

func formatProcLogDetailLines(logs []hoststats.ProcLog) []string {
	out := make([]string, 0, len(logs))
	for _, lg := range logs {
		out = append(out, formatProcLogStamp(lg.Time)+procLogStampSep+lg.Message)
	}
	return out
}

func writeCappedDetailList(b *strings.Builder, label string, items []string, remain *int) {
	if *remain < 1 {
		return
	}
	b.WriteString(label)
	b.WriteByte('\n')
	*remain--
	if len(items) == 0 {
		if *remain < 1 {
			return
		}
		b.WriteString(procDetailNone)
		b.WriteByte('\n')
		*remain--
		return
	}
	for i, item := range items {
		if *remain < 1 {
			return
		}
		left := len(items) - i
		if left > 1 && *remain == 1 {
			fmt.Fprintf(b, procDetailMoreFmt+"\n", left)
			*remain--
			return
		}
		fmt.Fprintf(b, procDetailItemFmt, item)
		*remain--
	}
}

func formatProcLogStamp(t time.Time) string {
	if t.IsZero() {
		return procLogStampEmpty
	}
	return t.Format(hoststats.ProcLogTimeLayout)
}
