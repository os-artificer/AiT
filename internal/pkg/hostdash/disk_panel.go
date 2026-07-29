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

// DiskPanel shows disk usage and IO rates.
type DiskPanel struct {
	text string
	hist []float64
}

// NewDiskPanel returns a disk panel.
func NewDiskPanel() *DiskPanel { return &DiskPanel{} }

// Title implements Panel.
func (p *DiskPanel) Title() string { return titleDISK }

// Update implements Panel.
func (p *DiskPanel) Update(s hoststats.Snapshot) {
	p.text = fmt.Sprintf("%s  %.1f%%\nr %s/s  w %s/s",
		s.Disk.Mount, s.Disk.UsedPercent,
		hoststats.FormatBytes(uint64(s.Disk.ReadBytesPerSec)),
		hoststats.FormatBytes(uint64(s.Disk.WriteBytesPerSec)))
	p.hist = appendSpark(p.hist, s.Disk.UsedPercent)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *DiskPanel) View(width, height int) string {
	cw, _ := contentBoxSize(width, height)
	body := titleStyle.Render(titleDISK) + "\n" + p.text + "\n" + renderSpark(p.hist, sparkInnerWidth(cw))
	return renderBox(body, width, height)
}
