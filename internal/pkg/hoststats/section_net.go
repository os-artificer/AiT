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

// NetSection formats network metrics.
type NetSection struct{}

// ID implements Section.
func (NetSection) ID() string { return IDNet }

// Human implements Section.
func (NetSection) Human(s Snapshot) string {
	return fmt.Sprintf("net %s: ↓%s/s ↑%s/s  errs %d/%d  drops %d/%d",
		s.Net.Name,
		FormatBytes(uint64(s.Net.RxBytesPerSec)),
		FormatBytes(uint64(s.Net.TxBytesPerSec)),
		s.Net.ErrIn, s.Net.ErrOut, s.Net.DropIn, s.Net.DropOut)
}

// JSONValue implements Section.
func (NetSection) JSONValue(s Snapshot) any { return s.Net }
