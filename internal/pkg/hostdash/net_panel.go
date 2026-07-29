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

// NetPanel shows network rates and drops.
type NetPanel struct {
	text string
	hist []float64
}

// NewNetPanel returns a network panel.
func NewNetPanel() *NetPanel { return &NetPanel{} }

// Title implements Panel.
func (p *NetPanel) Title() string { return titleNET }

// Update implements Panel.
func (p *NetPanel) Update(s hoststats.Snapshot) {
	rate := s.Net.RxBytesPerSec + s.Net.TxBytesPerSec
	p.text = fmt.Sprintf("%s\n↓%s/s ↑%s/s\nerrs %d/%d drops %d/%d",
		s.Net.Name,
		hoststats.FormatBytes(uint64(s.Net.RxBytesPerSec)),
		hoststats.FormatBytes(uint64(s.Net.TxBytesPerSec)),
		s.Net.ErrIn, s.Net.ErrOut, s.Net.DropIn, s.Net.DropOut)
	p.hist = appendSpark(p.hist, rate)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *NetPanel) View(width, height int) string {
	cw, _ := contentBoxSize(width, height)
	body := titleStyle.Render(titleNET) + "\n" + p.text + "\n" + renderSpark(p.hist, sparkInnerWidth(cw))
	return renderBox(body, width, height)
}
