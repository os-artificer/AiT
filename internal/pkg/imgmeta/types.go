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

// Package imgmeta extracts best-effort image metadata (EXIF, dims, GPS) for aiimg.
package imgmeta

import (
	"time"

	"github.com/evanoberholster/imagemeta/meta/exif"
)

// Info holds extracted image metadata for display.
// Address is filled by the caller after reverse geocode (formatters do not dial the network).
// Exif, when non-nil, is the structured decode used by FormatFullHuman / --all JSON.
type Info struct {
	File         string
	Format       string
	Width        int
	Height       int
	Make         string
	Model        string
	Software     string
	LensMake     string
	LensModel    string
	SerialNumber string
	DateTime     time.Time
	Address      string
	Latitude     *float64
	Longitude    *float64
	Altitude     *float64
	Exif         *exif.Exif
}

// HasGPS reports whether EXIF GPS coordinates are present.
func (i Info) HasGPS() bool {
	return i.Latitude != nil && i.Longitude != nil
}
