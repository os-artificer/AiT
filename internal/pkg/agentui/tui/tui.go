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

// Package tui provides the Bubble Tea agent UI (Phase 2 default on TTY).
package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type tickMsg struct{}

type splashModel struct {
	quitting bool
}

func (m splashModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m splashModel) View() string {
	if m.quitting {
		return ""
	}
	return "AiT Agent\n"
}

// RunLoop runs the agent task loop. On TTY it briefly uses alternate screen then delegates to agentcore.
func RunLoop(ctx context.Context, cfg agentcore.Config, initialGoal string) error {
	if shouldUseTUI(cfg.In, cfg.Out) {
		p := tea.NewProgram(splashModel{}, tea.WithContext(ctx), tea.WithOutput(cfg.Out), tea.WithInput(cfg.In))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("agent tui: %w", err)
		}
	}
	return agentcore.RunLoop(ctx, cfg, initialGoal)
}

func shouldUseTUI(in io.Reader, out io.Writer) bool {
	inFile, inOK := in.(*os.File)
	outFile, outOK := out.(*os.File)
	if !inOK || !outOK {
		return false
	}
	return term.IsTerminal(int(inFile.Fd())) && term.IsTerminal(int(outFile.Fd()))
}
