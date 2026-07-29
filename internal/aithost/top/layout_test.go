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

package top

import (
	"context"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

func TestPanelRowHeights_FitsTerminal(t *testing.T) {
	for _, h := range []int{4, 10, 16, 24, 40, 80} {
		topH, midH, procH := panelRowHeights(h)
		sum := topH + midH + procH + footerReserve
		if sum != h {
			t.Fatalf("h=%d top=%d mid=%d proc=%d sum=%d", h, topH, midH, procH, sum)
		}
		if topH < 1 || midH < 1 || procH < 1 {
			t.Fatalf("h=%d got non-positive row heights", h)
		}
	}
}

func TestView_NoVerticalOverflow(t *testing.T) {
	m := newModel(context.Background(), 0, hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{}))
	m.ready = true
	for _, h := range []int{16, 24, 40} {
		m.width = 80
		m.height = h
		for _, p := range m.panels {
			p.Update(hoststats.Snapshot{
				CPU:       hoststats.CPUStat{Percent: 10},
				Mem:       hoststats.MemStat{UsedPercent: 50, Used: 1, Total: 2},
				Processes: []hoststats.ProcStat{{PID: 1, Name: "init"}},
			})
		}
		view := m.View()
		if got := lipgloss.Height(view); got > h {
			t.Fatalf("h=%d viewHeight=%d > term\n%q", h, got, view)
		}
	}
}
