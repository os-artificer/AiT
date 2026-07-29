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

// DiskSection formats disk metrics.
type DiskSection struct{}

// ID implements Section.
func (DiskSection) ID() string { return IDDisk }

// Human implements Section.
func (DiskSection) Human(s Snapshot) string {
	return fmt.Sprintf("disk %s: %.1f%%  r %s/s  w %s/s",
		s.Disk.Mount, s.Disk.UsedPercent,
		FormatBytes(uint64(s.Disk.ReadBytesPerSec)),
		FormatBytes(uint64(s.Disk.WriteBytesPerSec)))
}

// JSONValue implements Section.
func (DiskSection) JSONValue(s Snapshot) any { return s.Disk }
