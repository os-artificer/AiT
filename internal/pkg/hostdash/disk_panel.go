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
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

// DiskPanel shows disk usage and IO rates.
type DiskPanel struct {
	value  string
	detail string
	histR  []float64
	histW  []float64
	times  []time.Time
}

// NewDiskPanel returns a disk panel.
func NewDiskPanel() *DiskPanel { return &DiskPanel{} }

// Title implements Panel.
func (p *DiskPanel) Title() string { return titleDISK }

// Update implements Panel.
func (p *DiskPanel) Update(s hoststats.Snapshot) {
	p.value = formatPercentValue(s.Disk.UsedPercent)
	readRate := diskReadPrefix + hoststats.FormatBytes(uint64(s.Disk.ReadBytesPerSec)) + rateUnit
	writeRate := diskWritePrefix + hoststats.FormatBytes(uint64(s.Disk.WriteBytesPerSec)) + rateUnit
	p.detail = fmt.Sprintf("%s  %s  %s",
		s.Disk.Mount,
		colorSparkSide(readRate, true),
		colorSparkSide(writeRate, false))

	p.histR, p.histW, p.times = appendSparkDual(
		p.histR, p.histW, p.times,
		s.Disk.ReadBytesPerSec, s.Disk.WriteBytesPerSec, s.Time)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *DiskPanel) View(width, height int) string {
	return renderDualSparkView(
		titleDISK, p.value, "", p.detail,
		p.histR, p.histW, p.times, width, height)
}
