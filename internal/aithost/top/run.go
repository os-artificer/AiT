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

// Package top implements the aithost live dashboard (default route and `aithost top`).
package top

import (
	"context"
	"fmt"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hostdash"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type sampleMsg struct {
	snap hoststats.Snapshot
	err  error
}

// Options configures the live dashboard.
type Options struct {
	Interval time.Duration
	Limit    int
	Sort     string
}

type model struct {
	ctx      context.Context
	interval time.Duration
	pq       hoststats.ProcessQuery
	col      *hoststats.Collector
	panels   []hostdash.Panel
	width    int
	height   int
	err      error
	ready    bool
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.sampleCmd(), m.tickCmd())
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		return m, tea.Batch(m.sampleCmd(), m.tickCmd())
	case sampleMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		for _, p := range m.panels {
			p.Update(msg.snap)
		}
		return m, nil
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n(press q to quit)\n", m.err)
	}
	if !m.ready && m.width == 0 {
		return "loading…"
	}
	w, h := m.width, m.height
	if w < 1 {
		w = fallbackWidth
	}
	if h < 1 {
		h = fallbackHeight
	}
	halfW := w / 2
	topH, midH, procH := panelRowHeights(h)

	cpu := m.panels[hostdash.PanelCPU].View(halfW, topH)
	mem := m.panels[hostdash.PanelMEM].View(w-halfW, topH)
	disk := m.panels[hostdash.PanelDISK].View(halfW, midH)
	net := m.panels[hostdash.PanelNET].View(w-halfW, midH)
	proc := m.panels[hostdash.PanelPROC].View(w, procH)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, cpu, mem),
		lipgloss.JoinHorizontal(lipgloss.Top, disk, net),
		proc,
		fmt.Sprintf(" q quit · interval %s · limit %d · sort %s", m.interval, m.pq.Limit, m.pq.Sort),
	)
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) sampleCmd() tea.Cmd {
	col, pq, ctx := m.col, m.pq, m.ctx
	return func() tea.Msg {
		snap, err := col.Sample(ctx, pq)
		return sampleMsg{snap: snap, err: err}
	}
}
