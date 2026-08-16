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
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// DefaultPanels returns the dashboard panels in layout order:
// CPU, MEM, DISK, NET, TCP, UDP, ARP, ICMP, PROC.
func DefaultPanels() []Panel {
	return []Panel{
		NewCPUPanel(),
		NewMemPanel(),
		NewDiskPanel(),
		NewNetPanel(),
		NewTCPPanel(),
		NewUDPPanel(),
		NewARPPanel(),
		NewICMPPanel(),
		NewProcPanel(),
	}
}

// ProcPanelRowCapacity is how many process rows fit in a PROC panel of outerH cells
// (including the rounded border). The result is at least 1.
func ProcPanelRowCapacity(outerH int) int {
	_, ch := contentBoxSize(minContentW, outerH)
	return procRowCapacity(ch)
}

// contentBoxSize converts outer terminal cells to lipgloss Width/Height
// (border is drawn outside those dimensions).
func contentBoxSize(outerW, outerH int) (w, h int) {
	w = outerW - boxBorder
	h = outerH - boxBorder
	if w < minContentW {
		w = minContentW
	}
	if h < minContentH {
		h = minContentH
	}
	return w, h
}

func procRowCapacity(contentH int) int {
	rows := contentH - procPanelTextLines
	if rows < 1 {
		return 1
	}
	return rows
}

func sparkInnerWidth(contentW int) int {
	inner := contentW - sparkBoxPad
	if inner < minContentW {
		inner = minContentW
	}
	return inner
}

func formatPercentValue(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

// renderBox paints body into an outer width x height cell box (border outside Width/Height).
func renderBox(body string, outerW, outerH int) string {
	cw, ch := contentBoxSize(outerW, outerH)
	return boxStyle.Width(cw).MaxWidth(outerW).Height(ch).MaxHeight(outerH).Render(body)
}

func countTextLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// formatPanelHeadline joins a panel type and primary value (e.g. "CPU   15.0%").
func formatPanelHeadline(title, value string) string {
	return title + panelHeadlineSep + value
}

// centerPanelLine pads s so it is centered in width display cells (ANSI-aware).
// When s is longer than width, s is returned unchanged.
func centerPanelLine(s string, width int) string {
	n := lipgloss.Width(s)
	if width < 1 || n >= width {
		return s
	}
	pad := width - n
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

func colorANSI(s, code string) string {
	if s == "" || code == "" {
		return s
	}
	return code + s + ansiReset
}

func colorSparkSide(s string, left bool) string {
	if left {
		return colorANSI(s, ansiIn)
	}
	return colorANSI(s, ansiOut)
}

func colorSparkUint(n uint64, left bool) string {
	return colorSparkSide(strconv.FormatUint(n, 10), left)
}

func colorInOutPair(in, out uint64) string {
	return colorSparkUint(in, true) + inOutSlash + colorSparkUint(out, false)
}

func formatInOutHeadline(inRate, outRate string) (plain, styled string) {
	inLbl := glyphIn + inRate
	outLbl := glyphOut + outRate
	plain = inLbl + sparkValueSep + outLbl
	styled = colorSparkSide(inLbl, true) + sparkValueSep + colorSparkSide(outLbl, false)
	return plain, styled
}

func yTickPaint(label, tint string) string {
	if tint != "" {
		return colorANSI(label, tint)
	}
	return axisStyle.Render(label)
}

// renderPanelHeadline returns a centered "TITLE   value" line with a bold title.
// Padding is based on the plain (unstyled) headline width so ANSI codes do not skew centering.
func renderPanelHeadline(title, value string, width int) string {
	return renderPanelHeadlineStyled(title, value, axisStyle.Render(value), width)
}

// renderPanelHeadlineStyled is renderPanelHeadline with a pre-colored value (plainValue is used for width).
func renderPanelHeadlineStyled(title, plainValue, styledValue string, width int) string {
	plain := formatPanelHeadline(title, plainValue)
	styled := titleStyle.Render(title) + panelHeadlineSep + styledValue
	n := utf8.RuneCountInString(plain)
	if width < 1 || n >= width {
		return styled
	}
	left := (width - n) / 2
	return strings.Repeat(" ", left) + styled
}

func sparkPanelBody(headline, detail string, inner int) string {
	if detail == "" {
		return headline
	}
	return headline + "\n" + centerPanelLine(detail, inner)
}

func renderDualSparkView(
	title, plain, styled, detail string,
	histL, histR []float64, times []time.Time,
	width, height int,
) string {
	cw, ch := contentBoxSize(width, height)
	inner := sparkInnerWidth(cw)
	headline := renderPanelHeadline(title, plain, inner)
	if styled != "" {
		headline = renderPanelHeadlineStyled(title, plain, styled, inner)
	}
	body := sparkPanelBody(headline, detail, inner)
	chart := renderSparkChartDual(
		histL, histR, times, inner, sparkCurveHeight(ch, countTextLines(body)), rateSparkScale())
	return renderBox(body+"\n"+chart, width, height)
}

func renderTintedSparkView(
	title, value, detail, tint string,
	hist []float64, times []time.Time,
	scale sparkScale, width, height int,
) string {
	cw, ch := contentBoxSize(width, height)
	inner := sparkInnerWidth(cw)
	body := sparkPanelBody(
		renderPanelHeadlineStyled(title, value, colorANSI(value, tint), inner),
		detail, inner)
	chart := renderSparkChartTinted(
		hist, times, inner, sparkCurveHeight(ch, countTextLines(body)), scale, tint)
	return renderBox(body+"\n"+chart, width, height)
}
