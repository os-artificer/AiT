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

import "fmt"

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

// FormatProcRow formats one process as a fixed-width table line (no trailing newline).
func FormatProcRow(p ProcStat) string {
	return fmt.Sprintf("%-8d %-20s %5.1f  %5.1f  %-8s  %-8s",
		p.PID, TruncateName(p.Name, ProcNameColWidth), p.CPUPercent, p.MemPercent,
		FormatBytes(uint64(p.DiskBytesPerSec)),
		FormatBytes(uint64(p.NetBytesPerSec)))
}
