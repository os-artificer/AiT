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

import (
	"strings"
)

// ProcSection formats the process table.
type ProcSection struct{}

// ID implements Section.
func (ProcSection) ID() string { return IDProc }

// Human implements Section.
func (ProcSection) Human(s Snapshot) string {
	var b strings.Builder
	b.WriteString(ProcTableHeader)
	b.WriteByte('\n')
	for _, p := range s.Processes {
		b.WriteString(FormatProcRow(p))
		b.WriteByte('\n')
	}
	return b.String()
}

// JSONValue implements Section.
func (ProcSection) JSONValue(s Snapshot) any { return s.Processes }
