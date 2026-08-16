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

// NetPanel shows network rates and drops.
type NetPanel struct {
	value  string
	styled string
	detail string
	histRx []float64
	histTx []float64
	times  []time.Time
}

// NewNetPanel returns a network panel.
func NewNetPanel() *NetPanel { return &NetPanel{} }

// Title implements Panel.
func (p *NetPanel) Title() string { return titleNET }

// Update implements Panel.
func (p *NetPanel) Update(s hoststats.Snapshot) {
	rx := hoststats.FormatBytes(uint64(s.Net.RxBytesPerSec)) + rateUnit
	tx := hoststats.FormatBytes(uint64(s.Net.TxBytesPerSec)) + rateUnit
	p.value, p.styled = formatInOutHeadline(rx, tx)
	p.detail = fmt.Sprintf("%s  errs %s drops %s",
		s.Net.Name,
		colorInOutPair(s.Net.ErrIn, s.Net.ErrOut),
		colorInOutPair(s.Net.DropIn, s.Net.DropOut))

	p.histRx, p.histTx, p.times = appendSparkDual(
		p.histRx, p.histTx, p.times,
		s.Net.RxBytesPerSec, s.Net.TxBytesPerSec, s.Time)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *NetPanel) View(width, height int) string {
	return renderDualSparkView(
		titleNET, p.value, p.styled, p.detail,
		p.histRx, p.histTx, p.times, width, height)
}
