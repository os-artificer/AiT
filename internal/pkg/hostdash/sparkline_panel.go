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
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

const (
	sparkCPU sparkKind = iota
	sparkMem
)

type sparkKind int

// SparklinePanel is a titled panel with a percent sparkline history.
type SparklinePanel struct {
	kind    sparkKind
	title   string
	history []float64
	times   []time.Time
	scale   sparkScale
	tint    string
	value   string
	detail  string
}

// NewCPUPanel returns the CPU sparkline panel.
func NewCPUPanel() *SparklinePanel {
	return newSparklinePanel(sparkCPU, titleCPU, ansiCPU)
}

// NewMemPanel returns the memory sparkline panel.
func NewMemPanel() *SparklinePanel {
	return newSparklinePanel(sparkMem, titleMEM, ansiMEM)
}

func newSparklinePanel(kind sparkKind, title, tint string) *SparklinePanel {
	return &SparklinePanel{
		kind:    kind,
		title:   title,
		history: make([]float64, 0, sparkLen),
		times:   make([]time.Time, 0, sparkLen),
		scale:   percentSparkScale(),
		tint:    tint,
	}
}

// Title implements Panel.
func (p *SparklinePanel) Title() string { return p.title }

// Update implements Panel. Refreshes percent lines and sparkline history from s.
func (p *SparklinePanel) Update(s hoststats.Snapshot) {
	var pct float64
	switch p.kind {
	case sparkCPU:
		pct = s.CPU.Percent
		p.value = formatPercentValue(pct)
		p.detail = ""
	default:
		pct = s.Mem.UsedPercent
		p.value = formatPercentValue(pct)
		p.detail = colorANSI(hoststats.FormatBytes(s.Mem.Used), ansiMEM) +
			memUsedSep + colorANSI(hoststats.FormatBytes(s.Mem.Total), ansiMEM)
	}

	p.history, p.times = appendSparkPoint(p.history, p.times, pct, s.Time)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *SparklinePanel) View(width, height int) string {
	return renderTintedSparkView(
		p.title, p.value, p.detail, p.tint,
		p.history, p.times, p.scale, width, height)
}
