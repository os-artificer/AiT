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
	"strings"
)

var (
	// ProcTableHeader is the default process table column header (no trailing newline).
	ProcTableHeader = FormatProcTableHeader(ProcLogColWidth)
)

// FormatBytes renders byte counts in human units (B, KB, MB, …).
func FormatBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := uint64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// FormatPacketRate renders a packets-per-second value for protocol panel headlines.
func FormatPacketRate(v float64) string {
	if v < 0 {
		v = 0
	}
	if v < 1000 {
		return fmt.Sprintf("%.0f", v)
	}
	const unit = 1000.0
	div, exp := unit, 0
	for n := v / unit; n >= unit && exp < 4; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", v/div, "KMGTPE"[exp])
}

// TruncateName shortens s to at most n runes for table columns, appending "..." when shortened.
func TruncateName(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-3]) + "..."
}

// FormatProcPeer renders listen ports or remote addresses for the PEER column.
// Servers (any Listen entry) show ":port×conns"; otherwise remotes as "host:port".
// The result is truncated to ProcPeerColWidth.
func FormatProcPeer(p ProcStat) string {
	lines := FormatProcPeerLines(p)
	if len(lines) == 0 {
		return ""
	}
	return TruncateName(strings.Join(lines, ","), ProcPeerColWidth)
}

// FormatProcPeerLines returns one full peer record per line for the PROC detail overlay.
// Listen entries win over remotes (same rule as FormatProcPeer). Empty when there are no peers.
func FormatProcPeerLines(p ProcStat) []string {
	switch {
	case len(p.Listen) > 0:
		out := make([]string, 0, len(p.Listen))
		for _, l := range p.Listen {
			out = append(out, fmt.Sprintf(":%d×%d", l.Port, l.Conns))
		}
		return out
	case len(p.Remotes) > 0:
		return append([]string(nil), p.Remotes...)
	default:
		return nil
	}
}

// FormatProcTableHeader returns the process table header with a LOG column of logW cells.
func FormatProcTableHeader(logW int) string {
	logW = normalizeProcLogWidth(logW)
	return fmt.Sprintf("%-*s %-*s %*s  %*s  %-*s  %-*s  %-*s",
		procPIDColWidth, "PID",
		ProcNameColWidth, "NAME",
		procMetricColWidth, "CPU%",
		procMetricColWidth, "MEM%",
		procDiskColWidth, "DISK/s",
		ProcPeerColWidth, "PEER",
		logW, "LOG")
}

// FormatProcLog returns the newest log line for the LOG column, flattened and truncated to width.
func FormatProcLog(p ProcStat, width int) string {
	width = normalizeProcLogWidth(width)
	msg := p.LogLine
	if msg == "" && len(p.Logs) > 0 {
		msg = p.Logs[len(p.Logs)-1].Message
	}
	return TruncateName(flattenLogMessage(msg), width)
}

// FormatProcRow formats one process as a fixed-width table line (no trailing newline).
func FormatProcRow(p ProcStat) string {
	return FormatProcRowWidth(p, ProcLogColWidth)
}

// FormatProcRowWidth formats one process row with a LOG column of logW cells.
func FormatProcRowWidth(p ProcStat, logW int) string {
	logW = normalizeProcLogWidth(logW)
	return fmt.Sprintf("%-*d %-*s %*.1f  %*.1f  %-*s  %-*s  %-*s",
		procPIDColWidth, p.PID,
		ProcNameColWidth, TruncateName(p.Name, ProcNameColWidth),
		procMetricColWidth, p.CPUPercent,
		procMetricColWidth, p.MemPercent,
		procDiskColWidth, FormatBytes(uint64(p.DiskBytesPerSec)),
		ProcPeerColWidth, FormatProcPeer(p),
		logW, FormatProcLog(p, logW))
}

func normalizeProcLogWidth(logW int) int {
	if logW < 1 {
		return ProcLogColWidth
	}
	return logW
}

func flattenLogMessage(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
