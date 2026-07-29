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

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

type sparkKind int

const (
	sparkCPU sparkKind = iota
	sparkMem
)

// SparklinePanel is a titled panel with a percent sparkline history.
type SparklinePanel struct {
	kind    sparkKind
	title   string
	history []float64
	line1   string
	line2   string
}

// NewCPUPanel returns the CPU sparkline panel.
func NewCPUPanel() *SparklinePanel {
	return &SparklinePanel{kind: sparkCPU, title: titleCPU, history: make([]float64, 0, sparkLen)}
}

// NewMemPanel returns the memory sparkline panel.
func NewMemPanel() *SparklinePanel {
	return &SparklinePanel{kind: sparkMem, title: titleMEM, history: make([]float64, 0, sparkLen)}
}

// Title implements Panel.
func (p *SparklinePanel) Title() string { return p.title }

// Update implements Panel. Refreshes percent lines and sparkline history from s.
func (p *SparklinePanel) Update(s hoststats.Snapshot) {
	var pct float64
	switch p.kind {
	case sparkCPU:
		pct = s.CPU.Percent
		p.line1 = fmt.Sprintf("%.1f%%", pct)
		p.line2 = ""
	default:
		pct = s.Mem.UsedPercent
		p.line1 = fmt.Sprintf("%.1f%%", pct)
		p.line2 = fmt.Sprintf("%s / %s", hoststats.FormatBytes(s.Mem.Used), hoststats.FormatBytes(s.Mem.Total))
	}
	p.history = appendSpark(p.history, pct)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *SparklinePanel) View(width, height int) string {
	cw, _ := contentBoxSize(width, height)
	body := titleStyle.Render(p.title) + "\n" + p.line1
	if p.line2 != "" {
		body += "\n" + p.line2
	}
	body += "\n" + renderSpark(p.history, sparkInnerWidth(cw))
	return renderBox(body, width, height)
}
