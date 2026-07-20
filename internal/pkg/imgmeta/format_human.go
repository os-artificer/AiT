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

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const humanLabelWidth = 13

// FormatHuman writes every locked field in canonical order to w.
// Missing string-like values print as "0"; missing GPS/altitude/dims print as zero.
func FormatHuman(w io.Writer, info Info) error {
	lat, lon := gpsOrZero(info)
	alt := altitudeOrZero(info)

	fields := []struct {
		label string
		value string
	}{
		{"File", displayOrZero(info.File)},
		{"Format", displayOrZero(info.Format)},
		{"Width", strconv.Itoa(info.Width)},
		{"Height", strconv.Itoa(info.Height)},
		{"Make", displayOrZero(info.Make)},
		{"Model", displayOrZero(info.Model)},
		{"Software", displayOrZero(info.Software)},
		{"LensMake", displayOrZero(info.LensMake)},
		{"LensModel", displayOrZero(info.LensModel)},
		{"SerialNumber", displayOrZero(info.SerialNumber)},
		{"DateTime", dateTimeOrZero(info.DateTime)},
		{"Location", displayOrZero(info.Address)},
		{"GPS", fmt.Sprintf("%.6f, %.6f", lat, lon)},
		{"Altitude", formatAltitude(alt) + " m"},
	}
	for _, f := range fields {
		if err := writeHumanField(w, f.label, f.value); err != nil {
			return err
		}
	}
	return nil
}

func writeHumanField(w io.Writer, label, value string) error {
	padded := label + ":"
	if len(padded) < humanLabelWidth {
		padded += strings.Repeat(" ", humanLabelWidth-len(padded))
	} else {
		padded += " "
	}
	_, err := fmt.Fprintf(w, "%s%s\n", padded, value)
	return err
}

func displayOrZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}

func dateTimeOrZero(t time.Time) string {
	if t.IsZero() {
		return "0"
	}
	return formatDateTime(t)
}

func gpsOrZero(info Info) (lat, lon float64) {
	if info.HasGPS() {
		return *info.Latitude, *info.Longitude
	}
	return 0, 0
}

func altitudeOrZero(info Info) float64 {
	if info.Altitude != nil {
		return *info.Altitude
	}
	return 0
}

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

func formatAltitude(v float64) string {
	// Limit decimals then trim trailing zeros (avoids float32→float64 noise).
	s := strconv.FormatFloat(v, 'f', 3, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}
