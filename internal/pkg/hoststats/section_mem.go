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

package hoststats

import "fmt"

// MemSection formats memory metrics.
type MemSection struct{}

// ID implements Section.
func (MemSection) ID() string { return IDMem }

// Human implements Section.
func (MemSection) Human(s Snapshot) string {
	return fmt.Sprintf("mem: %.1f%% (%s / %s)",
		s.Mem.UsedPercent, FormatBytes(s.Mem.Used), FormatBytes(s.Mem.Total))
}

// JSONValue implements Section.
func (MemSection) JSONValue(s Snapshot) any { return s.Mem }
