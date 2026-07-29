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

// Package hostdash provides Bubble Tea panel primitives for aithost top.
package hostdash

import (
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

const (
	sparkLen = 48
	// boxBorder cells added by RoundedBorder on each axis (outside Width/Height).
	boxBorder = 2

	titleCPU  = "CPU"
	titleMEM  = "MEM"
	titleDISK = "DISK"
	titleNET  = "NET"
	titlePROC = "PROC"

	// Panel layout indices into DefaultPanels().
	PanelCPU  = 0
	PanelMEM  = 1
	PanelDISK = 2
	PanelNET  = 3
	PanelPROC = 4
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
