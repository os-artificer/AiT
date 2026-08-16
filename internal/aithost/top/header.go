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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerLabelStyle = lipgloss.NewStyle().Faint(true)
	headerValueStyle = lipgloss.NewStyle().Bold(true)
	headerClockStyle = lipgloss.NewStyle().Faint(true)
)

// renderHostHeader paints a wrapped, left-aligned host header.
// Labels are faint, values bold. Datetime is right-aligned on the last line only when
// it shares that line with other groups and leftover cells remain.
// Returns the rendered text and the number of lines (at least 1).
func renderHostHeader(ident hoststats.HostIdent, memTotal uint64, clock time.Time, width int) (string, int) {
	groups := hoststats.HostHeaderGroups(ident, memTotal, clock)
	if len(groups) == 0 {
		return "", 1
	}

	lines := hoststats.WrapHostHeader(groups, width)
	styled := make([]string, 0, len(lines))
	for i, line := range lines {
		if i == len(lines)-1 {
			styled = append(styled, styleHeaderLastLine(line, width))
			continue
		}
		styled = append(styled, styleHeaderLine(line))
	}
	text := strings.Join(styled, "\n")
	n := len(lines)
	if n < 1 {
		n = 1
	}
	return text, n
}

func styleHeaderLastLine(line string, width int) string {
	groups := splitHeaderGroups(line)
	if len(groups) == 0 {
		return ""
	}

	last := groups[len(groups)-1]
	isClock := looksLikeHeaderClock(last)
	if !isClock || len(groups) == 1 {
		return styleHeaderLine(line)
	}

	leftPlain := strings.Join(groups[:len(groups)-1], hoststats.HeaderFieldSep)
	leftStyled := styleHeaderLine(leftPlain)
	clockStyled := headerClockStyle.Render(last)
	leftW := utf8.RuneCountInString(leftPlain)
	clockW := utf8.RuneCountInString(last)
	if width > 0 && leftW+clockW < width {
		pad := width - leftW - clockW
		return leftStyled + strings.Repeat(" ", pad) + clockStyled
	}
	return leftStyled + hoststats.HeaderFieldSep + clockStyled
}

func styleHeaderLine(line string) string {
	groups := splitHeaderGroups(line)
	parts := make([]string, 0, len(groups))
	for _, g := range groups {
		parts = append(parts, styleHeaderGroup(g))
	}
	return strings.Join(parts, hoststats.HeaderFieldSep)
}

func styleHeaderGroup(g string) string {
	if looksLikeHeaderClock(g) {
		return headerClockStyle.Render(g)
	}
	label, value, ok := strings.Cut(g, " ")
	if !ok || !strings.HasSuffix(label, ":") {
		return headerValueStyle.Render(g)
	}
	return headerLabelStyle.Render(label) + " " + headerValueStyle.Render(value)
}

func splitHeaderGroups(line string) []string {
	if line == "" {
		return nil
	}
	return strings.Split(line, hoststats.HeaderFieldSep)
}

func looksLikeHeaderClock(s string) bool {
	if len(s) != len(hoststats.HeaderClockLayout) {
		return false
	}
	_, err := time.Parse(hoststats.HeaderClockLayout, s)
	return err == nil
}
