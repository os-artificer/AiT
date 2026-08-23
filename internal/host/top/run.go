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

// Package top implements the aihost live dashboard (default route and `aihost top`).
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
	host     hoststats.HostIdent
	memTotal uint64
	clock    time.Time
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
		return m, m.sampleCmd()
	case tea.KeyMsg:
		if m.handleDashKey(msg.String()) {
			return m, nil
		}
		if msg.String() == keyQuit || msg.String() == keyCtrlC {
			return m, tea.Quit
		}
	case tickMsg:
		return m, tea.Batch(m.sampleCmd(), m.tickCmd())
	case sampleMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.host = msg.snap.Host
		m.memTotal = msg.snap.Mem.Total
		m.clock = msg.snap.Time
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
		return viewLoading
	}
	w, h := m.width, m.height
	if w < 1 {
		w = fallbackWidth
	}
	if h < 1 {
		h = fallbackHeight
	}
	header, headerH := renderHostHeader(m.host, m.memTotal, m.clock, w)

	halfW := w / 2
	qW := w / 4
	topH, midH, protoH, procH := panelRowHeights(h, headerH)

	cpu := m.panels[hostdash.PanelCPU].View(halfW, topH)
	mem := m.panels[hostdash.PanelMEM].View(w-halfW, topH)
	disk := m.panels[hostdash.PanelDISK].View(halfW, midH)
	net := m.panels[hostdash.PanelNET].View(w-halfW, midH)
	tcp := m.panels[hostdash.PanelTCP].View(qW, protoH)
	udp := m.panels[hostdash.PanelUDP].View(qW, protoH)
	arp := m.panels[hostdash.PanelARP].View(qW, protoH)
	icmp := m.panels[hostdash.PanelICMP].View(w-3*qW, protoH)
	proc := m.panels[hostdash.PanelPROC].View(w, procH)
	rowsLabel := m.procRowsLabel()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, cpu, mem),
		lipgloss.JoinHorizontal(lipgloss.Top, disk, net),
		lipgloss.JoinHorizontal(lipgloss.Top, tcp, udp, arp, icmp),
		proc,
		fmt.Sprintf(footerHintFmt, m.interval, rowsLabel, m.pq.Sort),
	)
}

func (m model) handleDashKey(key string) bool {
	proc, ok := m.panels[hostdash.PanelPROC].(*hostdash.ProcPanel)
	if !ok {
		return false
	}
	return proc.HandleKey(key)
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) sampleCmd() tea.Cmd {
	col, pq, ctx := m.col, m.sampleProcessQuery(), m.ctx
	return func() tea.Msg {
		snap, err := col.Sample(ctx, pq)
		return sampleMsg{snap: snap, err: err}
	}
}

func (m model) sampleProcessQuery() hoststats.ProcessQuery {
	pq := m.pq
	if proc, ok := m.panels[hostdash.PanelPROC].(*hostdash.ProcPanel); ok {
		pq.KeepPIDs = proc.SampleKeepPIDs()
	}
	return pq
}

func (m model) procRowsLabel() string {
	if pp, ok := m.panels[hostdash.PanelPROC].(*hostdash.ProcPanel); ok {
		if n := pp.ProcessCount(); n > 0 {
			return fmt.Sprintf("%d", n)
		}
	}
	if m.pq.All {
		return "all"
	}
	return fmt.Sprintf("%d", m.pq.Limit)
}
