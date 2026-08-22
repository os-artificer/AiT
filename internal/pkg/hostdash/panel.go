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

// Package hostdash provides Bubble Tea panel primitives for aihost top.
package hostdash

import (
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	axisStyle  = lipgloss.NewStyle().Faint(true)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

const (
	// Series colors (ANSI-256; always emitted so CI/non-TTY still distinguish).
	ansiIn    = "\x1b[38;5;45m"  // cyan — in / read / rx
	ansiOut   = "\x1b[38;5;214m" // orange — out / write / tx
	ansiBoth  = "\x1b[38;5;213m" // pink when overlapping
	ansiCPU   = "\x1b[38;5;117m" // CPU percent + curve + Y ticks
	ansiMEM   = "\x1b[38;5;141m" // MEM percent + curve + Y ticks
	ansiReset = "\x1b[0m"

	glyphIn       = "↓"
	glyphOut      = "↑"
	rateUnit      = "/s"
	inOutSlash    = "/"
	sparkValueSep = " "
	memUsedSep    = " / "

	panelHeadlineSep = "   "
	diskReadPrefix   = "r "
	diskWritePrefix  = "w "

	sparkLen = 48
	// boxBorder cells added by RoundedBorder on each axis (outside Width/Height).
	boxBorder = 2
	// sparkBoxPad is lipgloss Padding(0, 1) inside Width.
	sparkBoxPad = 2
	minContentW = 8
	minContentH = 3

	titleCPU  = "CPU"
	titleMEM  = "MEM"
	titleDISK = "DISK"
	titleNET  = "NET"
	titleTCP  = "TCP"
	titleUDP  = "UDP"
	titleARP  = "ARP"
	titleICMP = "ICMP"
	titlePROC = "PROC"

	// procPanelTextLines is the PROC title plus column header inside the box.
	procPanelTextLines = 2

	// Panel layout indices into DefaultPanels().
	PanelCPU  = 0
	PanelMEM  = 1
	PanelDISK = 2
	PanelNET  = 3
	PanelTCP  = 4
	PanelUDP  = 5
	PanelARP  = 6
	PanelICMP = 7
	PanelPROC = 8
)

// Panel renders one dashboard region from hoststats snapshots.
type Panel interface {
	Title() string
	Update(s hoststats.Snapshot)
	// View renders into an outer rectangle of width x height terminal cells
	// (including the rounded border). Callers must ensure joined panels do not
	// exceed the terminal size.
	View(width, height int) string
}
