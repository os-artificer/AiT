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

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

// ProcPanel shows the top process table.
type ProcPanel struct {
	rows []hoststats.ProcStat
}

// NewProcPanel returns a process list panel.
func NewProcPanel() *ProcPanel { return &ProcPanel{} }

// Title implements Panel.
func (p *ProcPanel) Title() string { return titlePROC }

// Update implements Panel.
func (p *ProcPanel) Update(s hoststats.Snapshot) {
	p.rows = append([]hoststats.ProcStat(nil), s.Processes...)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *ProcPanel) View(width, height int) string {
	_, ch := contentBoxSize(width, height)
	var b strings.Builder
	b.WriteString(titleStyle.Render("PROC (top by CPU>MEM>NET>DISK)"))
	b.WriteByte('\n')
	b.WriteString(hoststats.ProcTableHeader)
	b.WriteByte('\n')
	maxRows := ch - 3
	if maxRows < 1 {
		maxRows = 1
	}
	for i, row := range p.rows {
		if i >= maxRows {
			break
		}
		b.WriteString(hoststats.FormatProcRow(row))
		b.WriteByte('\n')
	}
	return renderBox(strings.TrimRight(b.String(), "\n"), width, height)
}
