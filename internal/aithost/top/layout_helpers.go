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
)

// panelRowHeights splits termH into top/mid/proc row heights plus footerReserve.
// topH+midH+procH+footerReserve == termH when termH >= 4; each panel row is at least 1.
func panelRowHeights(termH int) (topH, midH, procH int) {
	avail := termH - footerReserve
	if avail < 3 {
		return 1, 1, 1
	}

	topH = avail / 4
	midH = avail / 4
	if topH < 1 {
		topH = 1
	}
	if midH < 1 {
		midH = 1
	}
	procH = avail - topH - midH
	if procH >= 1 {
		return topH, midH, procH
	}

	procH = 1
	rem := avail - procH
	topH = rem / 2
	if topH < 1 {
		topH = 1
	}
	midH = rem - topH
	if midH < 1 {
		midH = 1
		topH = avail - midH - procH
	}
	return topH, midH, procH
}
