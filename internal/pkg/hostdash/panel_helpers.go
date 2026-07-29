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

import "strings"

// DefaultPanels returns the five dashboard panels in layout order:
// CPU, MEM, DISK, NET, PROC.
func DefaultPanels() []Panel {
	return []Panel{
		NewCPUPanel(),
		NewMemPanel(),
		NewDiskPanel(),
		NewNetPanel(),
		NewProcPanel(),
	}
}

// contentBoxSize converts outer terminal cells to lipgloss Width/Height
// (border is drawn outside those dimensions).
func contentBoxSize(outerW, outerH int) (w, h int) {
	w = outerW - boxBorder
	h = outerH - boxBorder
	if w < 8 {
		w = 8
	}
	if h < 3 {
		h = 3
	}
	return w, h
}

func sparkInnerWidth(contentW int) int {
	// Padding(0, 1) eats 2 cells inside Width.
	inner := contentW - 2
	if inner < 8 {
		inner = 8
	}
	return inner
}

func appendSpark(hist []float64, v float64) []float64 {
	hist = append(hist, v)
	if len(hist) > sparkLen {
		return hist[len(hist)-sparkLen:]
	}
	return hist
}

// renderBox paints body into an outer width x height cell box (border outside Width/Height).
func renderBox(body string, outerW, outerH int) string {
	cw, ch := contentBoxSize(outerW, outerH)
	return boxStyle.Width(cw).Height(ch).MaxHeight(ch).Render(body)
}

func renderSpark(hist []float64, width int) string {
	if width <= 0 {
		return ""
	}
	bars := []rune("▁▂▃▄▅▆▇█")
	if len(hist) == 0 {
		return strings.Repeat(" ", width)
	}
	start := 0
	if len(hist) > width {
		start = len(hist) - width
	}
	slice := hist[start:]
	max := 0.0
	for _, v := range slice {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		max = 1
	}
	var b strings.Builder
	for _, v := range slice {
		idx := int(v / max * float64(len(bars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		b.WriteRune(bars[idx])
	}
	for b.Len() < width {
		b.WriteByte(' ')
	}
	return b.String()
}
