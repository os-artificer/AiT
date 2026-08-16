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

const (
	footerReserve  = 1
	fallbackWidth  = 80
	fallbackHeight = 24
	minPanelRows   = 4
	sparkRowShare  = 5
	keyQuit        = "q"
	keyCtrlC       = "ctrl+c"
	viewLoading    = "loading…"
	footerHintFmt  = " q quit · / search · enter detail · esc back · x kill · interval %s · rows %s · sort %s"
)

// panelRowHeights splits termH into top/mid/proto/proc row heights plus headerH and footer.
// topH+midH+protoH+procH+headerH+footerReserve == termH when termH >= headerH+footerReserve+minPanelRows;
// each panel row is at least 1. headerH is clamped to at least 1.
func panelRowHeights(termH, headerH int) (topH, midH, protoH, procH int) {
	if headerH < 1 {
		headerH = 1
	}
	avail := termH - headerH - footerReserve
	if avail < minPanelRows {
		return 1, 1, 1, 1
	}

	topH = avail / sparkRowShare
	midH = avail / sparkRowShare
	protoH = avail / sparkRowShare
	if topH < 1 {
		topH = 1
	}
	if midH < 1 {
		midH = 1
	}
	if protoH < 1 {
		protoH = 1
	}
	procH = avail - topH - midH - protoH
	if procH >= 1 {
		return topH, midH, protoH, procH
	}

	return splitThreePlusProc(avail)
}

func splitThreePlusProc(avail int) (topH, midH, protoH, procH int) {
	procH = 1
	rem := avail - procH
	topH = rem / 3
	if topH < 1 {
		topH = 1
	}
	midH = rem / 3
	if midH < 1 {
		midH = 1
	}
	protoH = rem - topH - midH
	if protoH >= 1 {
		return topH, midH, protoH, procH
	}
	protoH = 1
	midH = rem - topH - protoH
	if midH >= 1 {
		return topH, midH, protoH, procH
	}
	midH = 1
	topH = rem - midH - protoH
	return topH, midH, protoH, procH
}
