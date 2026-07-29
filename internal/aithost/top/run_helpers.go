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

package top

import (
	"context"
	"io"
	"time"

	"github.com/os-artificer/AiT/internal/aithost/shared"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/hostdash"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the live dashboard. Requires a TTY on stdout.
// Returns an errhint when stdout is not a terminal, or a Bubble Tea run error.
func Run(ctx context.Context, stdout, _ io.Writer, opts Options) error {
	if !shared.WriterIsTerminal(stdout) {
		return errhint.WithTryMsg(
			"top requires a TTY",
			"The live dashboard needs an interactive terminal.",
			toolname.AIHost+" snapshot --json",
			toolname.AIHost+" proc",
			toolname.AIHost+" --help",
		)
	}
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	pq := hoststats.NormalizeProcessQuery(hoststats.ProcessQuery{
		Limit: opts.Limit,
		Sort:  opts.Sort,
	})
	m := newModel(ctx, opts.Interval, pq)
	p := tea.NewProgram(m, tea.WithContext(ctx), tea.WithOutput(stdout), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(ctx context.Context, interval time.Duration, pq hoststats.ProcessQuery) model {
	return model{
		ctx:      ctx,
		interval: interval,
		pq:       pq,
		col:      hoststats.NewCollector(),
		panels:   hostdash.DefaultPanels(),
		width:    fallbackWidth,
		height:   fallbackHeight,
	}
}
