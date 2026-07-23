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

package imgmeta

import "time"

// PatchFields holds optional EXIF field updates for Patch.
// Nil pointers are left unchanged.
type PatchFields struct {
	Make         *string
	Model        *string
	Software     *string
	LensMake     *string
	LensModel    *string
	SerialNumber *string
	DateTime     *time.Time
	Latitude     *float64
	Longitude    *float64
	Altitude     *float64
}

// ClearMask selects which writable EXIF fields to remove for Clear.
type ClearMask struct {
	Make         bool
	Model        bool
	Software     bool
	LensMake     bool
	LensModel    bool
	SerialNumber bool
	DateTime     bool
	GPS          bool
	Altitude     bool
}

// Empty reports whether no field is selected.
func (m ClearMask) Empty() bool {
	return !m.Make && !m.Model && !m.Software && !m.LensMake && !m.LensModel &&
		!m.SerialNumber && !m.DateTime && !m.GPS && !m.Altitude
}

// AllClearMask returns a ClearMask with every clearable field enabled.
func AllClearMask() ClearMask {
	return ClearMask{
		Make: true, Model: true, Software: true,
		LensMake: true, LensModel: true, SerialNumber: true,
		DateTime: true, GPS: true, Altitude: true,
	}
}

// Empty reports whether the patch sets no fields.
func (p PatchFields) Empty() bool {
	return p.Make == nil && p.Model == nil && p.Software == nil &&
		p.LensMake == nil && p.LensModel == nil && p.SerialNumber == nil &&
		p.DateTime == nil && p.Latitude == nil && p.Longitude == nil && p.Altitude == nil
}

// HasGPS reports whether both latitude and longitude are set for writing.
func (p PatchFields) HasGPS() bool {
	return p.Latitude != nil && p.Longitude != nil
}
