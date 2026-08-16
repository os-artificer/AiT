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

// CPUSection formats CPU metrics.
type CPUSection struct{}

// ID implements Section.
func (CPUSection) ID() string { return IDCPU }

// Human implements Section.
func (CPUSection) Human(s Snapshot) string {
	return fmt.Sprintf("cpu: %.1f%%", s.CPU.Percent)
}

// JSONValue implements Section.
func (CPUSection) JSONValue(s Snapshot) any { return s.CPU }
