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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var (
	brailleDotBits = [brailleRowsPerCell][brailleColsPerCell]byte{
		{0x01, 0x08},
		{0x02, 0x10},
		{0x04, 0x20},
		{0x40, 0x80},
	}
)

const (
	brailleColsPerCell = 2
	brailleRowsPerCell = 4
	brailleBase        = 0x2800
	sparkYGutter       = 1
	sparkMinCurveW     = 4
	sparkXClock        = "15:04:05"
)

// sparkScale configures Y-axis range and tick labels for a spark chart.
type sparkScale struct {
	// FixedMax, when > 0, is the Y-axis maximum (e.g. 100 for percent).
	// When 0, the maximum is taken from the visible history (or 1 if empty).
	FixedMax float64
	FormatY  func(float64) string
}

func percentSparkScale() sparkScale {
	return sparkScale{FixedMax: 100, FormatY: formatPercentTick}
}

func rateSparkScale() sparkScale {
	return sparkScale{FormatY: formatRateTick}
}

// renderSparkChart draws a Braille curve with left Y ticks and a bottom X time axis.
// width/height are the full chart region (including axis gutters).
func renderSparkChart(hist []float64, times []time.Time, width, height int, scale sparkScale) string {
	return renderSparkChartTinted(hist, times, width, height, scale, "")
}

func renderSparkChartTinted(hist []float64, times []time.Time, width, height int, scale sparkScale, tint string) string {
	return renderSparkChartSeries([][]float64{hist}, times, width, height, scale, tint)
}

// renderSparkChartSeries overlays one or more Braille series on a shared Y scale.
// tint, when non-empty, colors the curve and left Y ticks (X-axis stays faint).
func renderSparkChartSeries(series [][]float64, times []time.Time, width, height int, scale sparkScale, tint string) string {
	if width < 1 || height < 1 {
		return ""
	}

	yMax := scaleYMaxSeries(series, scale)
	yW := yAxisWidth(scale, yMax)
	curveH, needX := sparkCurveRegion(height)
	plotW := width - yW - sparkYGutter
	cellW := sparkCellWidth()
	curveW := plotW / cellW
	if curveW < sparkMinCurveW {
		return renderSparkCurvesTinted(series, width, height, yMax, tint)
	}

	curve := renderSparkCurvesTinted(series, curveW, curveH, yMax, tint)
	curveLines := padCurveLines(curve, curveH, plotW)

	primary := primarySeries(series)
	var b strings.Builder
	for row := 0; row < curveH; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		tick := padLeft(yTickLabel(scale, yMax, row, curveH), yW)
		line := yTickPaint(tick, tint)
		line += strings.Repeat(" ", sparkYGutter)
		line += clampDisplayWidth(curveLines[row], plotW)
		b.WriteString(clampDisplayWidth(line, width))
	}
	if needX {
		b.WriteByte('\n')
		xline := strings.Repeat(" ", yW+sparkYGutter)
		xline += axisStyle.Render(formatXAxis(times, histVisibleStart(primary, curveW), plotW))
		b.WriteString(clampDisplayWidth(xline, width))
	}
	return fitChartLines(b.String(), height)
}

// renderSparkChartDual draws two series with independent left/right Y scales and colors.
// Left axis/histL is the first series (DISK read / NET rx); right is the second (write / tx).
func renderSparkChartDual(histL, histR []float64, times []time.Time, width, height int, scale sparkScale) string {
	if width < 1 || height < 1 {
		return ""
	}

	yMaxL := scaleYMax(histL, scale)
	yMaxR := scaleYMax(histR, scale)
	leftW := yAxisWidth(scale, yMaxL)
	rightW := yAxisWidth(scale, yMaxR)
	curveH, needX := sparkCurveRegion(height)
	plotW := width - leftW - rightW - 2*sparkYGutter
	cellW := sparkCellWidth()
	curveW := plotW / cellW
	if curveW < sparkMinCurveW {
		return encodeColoredBrailleGrids(
			brailleGrid(histL, width, height, yMaxL),
			brailleGrid(histR, width, height, yMaxR),
			width, height)
	}

	curve := encodeColoredBrailleGrids(
		brailleGrid(histL, curveW, curveH, yMaxL),
		brailleGrid(histR, curveW, curveH, yMaxR),
		curveW, curveH)
	curveLines := padCurveLines(curve, curveH, plotW)

	primary := histL
	if len(primary) == 0 {
		primary = histR
	}
	var b strings.Builder
	for row := 0; row < curveH; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		left := colorSparkSide(padLeft(yTickLabel(scale, yMaxL, row, curveH), leftW), true)
		right := colorSparkSide(padRight(yTickLabel(scale, yMaxR, row, curveH), rightW), false)
		line := left + strings.Repeat(" ", sparkYGutter)
		line += clampDisplayWidth(curveLines[row], plotW)
		line += strings.Repeat(" ", sparkYGutter) + right
		b.WriteString(clampDisplayWidth(line, width))
	}
	if needX {
		b.WriteByte('\n')
		xline := strings.Repeat(" ", leftW+sparkYGutter)
		xline += axisStyle.Render(formatXAxis(times, histVisibleStart(primary, curveW), plotW))
		xline += strings.Repeat(" ", sparkYGutter+rightW)
		b.WriteString(clampDisplayWidth(xline, width))
	}
	return fitChartLines(b.String(), height)
}

// fitChartLines keeps the last line (X-axis) when the chart is taller than maxLines.
func fitChartLines(chart string, maxLines int) string {
	if maxLines < 1 || chart == "" {
		return chart
	}
	lines := strings.Split(chart, "\n")
	if len(lines) <= maxLines {
		return chart
	}
	x := lines[len(lines)-1]
	curve := lines[:len(lines)-1]
	keep := maxLines - 1
	if keep < 0 {
		keep = 0
	}
	if len(curve) > keep {
		curve = curve[:keep]
	}
	return strings.Join(append(curve, x), "\n")
}

// renderSparkCurve draws hist as a Braille polyline of width×height cells.
// Y is normalized to yMax (yMax ≤ 0 treated as 1). Empty hist yields blank rows.
func renderSparkCurve(hist []float64, width, height int, yMax float64) string {
	return renderSparkCurves([][]float64{hist}, width, height, yMax)
}

func renderSparkCurves(series [][]float64, width, height int, yMax float64) string {
	return renderSparkCurvesTinted(series, width, height, yMax, "")
}

func renderSparkCurvesTinted(series [][]float64, width, height int, yMax float64, tint string) string {
	if width < 1 || height < 1 {
		return ""
	}
	empty := true
	for _, hist := range series {
		if len(hist) > 0 {
			empty = false
			break
		}
	}
	if empty {
		return blankSparkRows(width, height)
	}
	yMax = sparkYFloor(yMax)

	grid := make([]byte, width*height)
	for _, hist := range series {
		paintSparkPolyline(grid, hist, width, height, yMax)
	}
	return encodeBrailleGrid(grid, width, height, tint)
}

func sparkYFloor(yMax float64) float64 {
	if yMax <= 0 {
		return 1
	}
	return yMax
}

func sparkCurveRegion(height int) (curveH int, needX bool) {
	needX = height >= 2
	curveH = height
	if needX {
		curveH = height - 1
	}
	return curveH, needX
}

func padCurveLines(curve string, curveH, plotW int) []string {
	lines := strings.Split(curve, "\n")
	blank := strings.Repeat(" ", plotW)
	for len(lines) < curveH {
		lines = append(lines, blank)
	}
	return lines
}

func brailleGrid(hist []float64, width, height int, yMax float64) []byte {
	grid := make([]byte, width*height)
	paintSparkPolyline(grid, hist, width, height, sparkYFloor(yMax))
	return grid
}

func encodeColoredBrailleGrids(gridL, gridR []byte, width, height int) string {
	var b strings.Builder
	b.Grow(height * (width*(utf8.UTFMax+16) + 1))
	for row := 0; row < height; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		off := row * width
		for col := 0; col < width; col++ {
			left, right := brailleCellBits(gridL, gridR, off+col)
			b.WriteString(coloredBrailleCell(left, right))
		}
	}
	return b.String()
}

func brailleCellBits(gridL, gridR []byte, i int) (left, right byte) {
	if i < len(gridL) {
		left = gridL[i]
	}
	if i < len(gridR) {
		right = gridR[i]
	}
	return left, right
}

func brailleGlyph(bits byte) string {
	return string(rune(brailleBase) + rune(bits))
}

func coloredBrailleCell(left, right byte) string {
	switch {
	case left == 0 && right == 0:
		return strings.Repeat(" ", sparkCellWidth())
	case right == 0:
		return colorANSI(brailleGlyph(left), ansiIn)
	case left == 0:
		return colorANSI(brailleGlyph(right), ansiOut)
	default:
		return colorANSI(brailleGlyph(left|right), ansiBoth)
	}
}

func paintSparkPolyline(grid []byte, hist []float64, width, height int, yMax float64) {
	if len(hist) == 0 {
		return
	}
	start := 0
	if len(hist) > width {
		start = len(hist) - width
	}
	slice := hist[start:]

	dotW := width * brailleColsPerCell
	dotH := height * brailleRowsPerCell
	n := len(slice)
	for i := 0; i < n; i++ {
		x0 := sampleDotX(i, n, dotW)
		y0 := sampleDotY(slice[i], yMax, dotH)
		if i+1 >= n {
			setBrailleDot(grid, width, height, x0, y0)
			continue
		}
		x1 := sampleDotX(i+1, n, dotW)
		y1 := sampleDotY(slice[i+1], yMax, dotH)
		drawBrailleLine(grid, width, height, x0, y0, x1, y1)
	}
}

func sparkCurveHeight(contentH, textLines int) int {
	h := contentH - textLines
	if h < 1 {
		h = 1
	}
	return h
}

func appendSparkPoint(hist []float64, times []time.Time, v float64, t time.Time) ([]float64, []time.Time) {
	if t.IsZero() {
		t = time.Now()
	}
	hist = append(hist, v)
	times = append(times, t)
	if len(hist) > sparkLen {
		hist = hist[len(hist)-sparkLen:]
		times = times[len(times)-sparkLen:]
	}
	return hist, times
}

func appendSparkDual(histA, histB []float64, times []time.Time, a, b float64, t time.Time) ([]float64, []float64, []time.Time) {
	if t.IsZero() {
		t = time.Now()
	}
	histA = append(histA, a)
	histB = append(histB, b)
	times = append(times, t)
	if len(histA) > sparkLen {
		histA = histA[len(histA)-sparkLen:]
		histB = histB[len(histB)-sparkLen:]
		times = times[len(times)-sparkLen:]
	}
	return histA, histB, times
}

func blankSparkRows(width, height int) string {
	row := strings.Repeat(" ", width*sparkCellWidth())
	lines := make([]string, height)
	for i := range lines {
		lines[i] = row
	}
	return strings.Join(lines, "\n")
}

func scaleYMax(hist []float64, scale sparkScale) float64 {
	return scaleYMaxSeries([][]float64{hist}, scale)
}

func scaleYMaxSeries(series [][]float64, scale sparkScale) float64 {
	if scale.FixedMax > 0 {
		return scale.FixedMax
	}
	max := 0.0
	for _, hist := range series {
		if m := sparkMax(hist); m > max {
			max = m
		}
	}
	if max <= 0 {
		return 1
	}
	return max
}

func primarySeries(series [][]float64) []float64 {
	for _, hist := range series {
		if len(hist) > 0 {
			return hist
		}
	}
	if len(series) > 0 {
		return series[0]
	}
	return nil
}

func sparkMax(vals []float64) float64 {
	max := 0.0
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}

func yAxisWidth(scale sparkScale, yMax float64) int {
	format := scale.FormatY
	if format == nil {
		format = formatPlainTick
	}
	w := 0
	for _, v := range []float64{yMax, yMax / 2, 0} {
		n := utf8.RuneCountInString(format(v))
		if n > w {
			w = n
		}
	}
	if w < 3 {
		w = 3
	}
	return w
}

func yTickLabel(scale sparkScale, yMax float64, row, curveH int) string {
	format := scale.FormatY
	if format == nil {
		format = formatPlainTick
	}
	switch {
	case row == 0:
		return format(yMax)
	case curveH >= 3 && row == curveH/2:
		return format(yMax / 2)
	case row == curveH-1:
		return format(0)
	default:
		return ""
	}
}

func formatPercentTick(v float64) string {
	return fmt.Sprintf("%.0f%%", v)
}

func formatRateTick(v float64) string {
	if v < 0 {
		v = 0
	}
	return hoststats.FormatBytes(uint64(v)) + "/s"
}

func formatPlainTick(v float64) string {
	return fmt.Sprintf("%.0f", v)
}

func formatXAxis(times []time.Time, startIdx, curveW int) string {
	if curveW < 1 {
		return ""
	}
	if len(times) == 0 {
		return strings.Repeat(" ", curveW)
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(times) {
		startIdx = len(times) - 1
	}
	vis := times[startIdx:]
	labelW := len(sparkXClock)
	if curveW < labelW {
		s := vis[len(vis)-1].Format(sparkXClock)
		if len(s) > curveW {
			s = s[len(s)-curveW:]
		}
		return s + strings.Repeat(" ", curveW-len(s))
	}

	n := xAxisLabelCount(len(vis), curveW, labelW)
	buf := []rune(strings.Repeat(" ", curveW))
	if n == 1 {
		placeRunes(buf, 0, []rune(vis[len(vis)-1].Format(sparkXClock)))
		return string(buf)
	}
	for i := 0; i < n; i++ {
		idx := i * (len(vis) - 1) / (n - 1)
		at := i * (curveW - labelW) / (n - 1)
		placeRunes(buf, at, []rune(vis[idx].Format(sparkXClock)))
	}
	return string(buf)
}

func xAxisLabelCount(samples, curveW, labelW int) int {
	if samples < 1 || curveW < labelW {
		return 0
	}
	gap := 1
	maxN := (curveW + gap) / (labelW + gap)
	if maxN < 1 {
		return 1
	}
	if maxN > samples {
		maxN = samples
	}
	if maxN > 1 {
		span := curveW - labelW
		step := span / (maxN - 1)
		if step < labelW+gap {
			maxN = span/(labelW+gap) + 1
		}
	}
	if maxN < 1 {
		return 1
	}
	if maxN > samples {
		return samples
	}
	return maxN
}

func histVisibleStart(hist []float64, curveW int) int {
	if len(hist) > curveW {
		return len(hist) - curveW
	}
	return 0
}

func placeRunes(dst []rune, at int, src []rune) {
	for i, r := range src {
		if at+i < 0 || at+i >= len(dst) {
			continue
		}
		dst[at+i] = r
	}
}

func padLeft(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return strings.Repeat(" ", width-n) + s
}

func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func sampleDotX(i, n, dotW int) int {
	if n <= 1 || dotW <= 1 {
		return 0
	}
	return i * (dotW - 1) / (n - 1)
}

func sampleDotY(v, max float64, dotH int) int {
	if dotH <= 1 {
		return 0
	}
	if v < 0 {
		v = 0
	}
	if v > max {
		v = max
	}
	level := int(v / max * float64(dotH-1))
	if level < 0 {
		level = 0
	}
	if level >= dotH {
		level = dotH - 1
	}
	return (dotH - 1) - level
}

func drawBrailleLine(grid []byte, cellsW, cellsH, x0, y0, x1, y1 int) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	x, y := x0, y0
	for {
		setBrailleDot(grid, cellsW, cellsH, x, y)
		if x == x1 && y == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
}

func setBrailleDot(grid []byte, cellsW, cellsH, x, y int) {
	dotW := cellsW * brailleColsPerCell
	dotH := cellsH * brailleRowsPerCell
	if x < 0 || y < 0 || x >= dotW || y >= dotH {
		return
	}
	cellX := x / brailleColsPerCell
	cellY := y / brailleRowsPerCell
	localX := x % brailleColsPerCell
	localY := y % brailleRowsPerCell
	idx := cellY*cellsW + cellX
	grid[idx] |= brailleDotBits[localY][localX]
}

func encodeBrailleGrid(grid []byte, width, height int, tint string) string {
	empty := strings.Repeat(" ", sparkCellWidth())
	var b strings.Builder
	b.Grow(height * (width*(utf8.UTFMax+16) + 1))
	for row := 0; row < height; row++ {
		if row > 0 {
			b.WriteByte('\n')
		}
		off := row * width
		for col := 0; col < width; col++ {
			b.WriteString(tintedBrailleCell(grid[off+col], tint, empty))
		}
	}
	return b.String()
}

func tintedBrailleCell(mask byte, tint, empty string) string {
	if mask == 0 {
		return empty
	}
	cell := brailleGlyph(mask)
	if tint == "" {
		return cell
	}
	return colorANSI(cell, tint)
}

func sparkCellWidth() int {
	w := lipgloss.Width("⣿")
	if w < 1 {
		return 1
	}
	return w
}

func clampDisplayWidth(s string, w int) string {
	if w < 1 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(w).Width(w).Render(s)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
