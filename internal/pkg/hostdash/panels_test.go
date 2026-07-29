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
)

func TestDefaultPanels_IncludesProc(t *testing.T) {
	panels := DefaultPanels()
	found := false
	for _, p := range panels {
		if p.Title() == "PROC" {
			found = true
			p.Update(hoststats.Snapshot{
				Processes: []hoststats.ProcStat{{PID: 1, Name: "init", CPUPercent: 1}},
			})
			view := p.View(80, 12)
			if !strings.Contains(view, "PROC") || !strings.Contains(view, "PID") {
				t.Fatalf("view=%q", view)
			}
			if !strings.Contains(view, "init") {
				t.Fatalf("missing process row: %q", view)
			}
		}
	}
	if !found {
		t.Fatal("ProcPanel missing")
	}
}

func TestRenderSpark(t *testing.T) {
	s := renderSpark([]float64{0, 50, 100}, 3)
	if len([]rune(s)) < 3 {
		t.Fatalf("%q", s)
	}
}
