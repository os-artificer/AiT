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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

// AH: bordered panels joined horizontally must not exceed terminal width.
func TestPanelView_NoHorizontalOverflow(t *testing.T) {
	termW := 212
	half := termW / 2
	cpu := NewCPUPanel()
	mem := NewMemPanel()
	cpu.Update(hoststats.Snapshot{CPU: hoststats.CPUStat{Percent: 10.6}})
	mem.Update(hoststats.Snapshot{Mem: hoststats.MemStat{UsedPercent: 66.4, Used: 1, Total: 2}})
	left := cpu.View(half, 14)
	right := mem.View(termW-half, 14)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	for i, line := range strings.Split(row, "\n") {
		w := lipgloss.Width(line)
		if w > termW {
			t.Fatalf("row[%d] width=%d > termW=%d", i, w, termW)
		}
	}
	if !strings.Contains(left, "CPU") || !strings.Contains(left, "10.6%") {
		t.Fatalf("cpu panel missing content: %q", left)
	}
	if !strings.Contains(right, "MEM") || !strings.Contains(right, "66.4%") {
		t.Fatalf("mem panel missing content: %q", right)
	}
}
