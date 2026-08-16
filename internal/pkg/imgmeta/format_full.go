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
	"encoding/json"
	"fmt"
	"io"

	"github.com/evanoberholster/imagemeta/meta/exif"
)

// Canonical TIFF/EXIF IFD section order for --all human output.
var fullSectionOrder = []string{"IFD0", "ExifIFD", "GPS", "IFD1", "IFD2", "MakerNote"}

// Root-level exif.Exif fields printed before IFD sections.
var fullRootKeys = []string{"ImageType", "CameraSerial", "CameraMakeID"}

const fullExifHeader = "--- EXIF ---"

// FormatFullHuman writes the structured EXIF dump (IFD tree) to w as Field/Value tables.
// ex may be nil (header only). Absent/empty fields are skipped (unlike locked summary zeros).
// String values that parse as JSON, XML, or YAML structure are flattened; otherwise the raw string is shown.
func FormatFullHuman(w io.Writer, ex *exif.Exif) error {
	if _, err := fmt.Fprintln(w, fullExifHeader); err != nil {
		return err
	}
	if ex == nil {
		return nil
	}

	raw, err := json.Marshal(ex)
	if err != nil {
		return err
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return err
	}

	rootRows := collectRootRows(top)
	if err := writeNamedTable(w, "", rootRows); err != nil {
		return err
	}

	for _, sec := range fullSectionOrder {
		val, ok := top[sec]
		if !ok || jsonValueEmpty(val) {
			continue
		}
		if err := writeNamedTable(w, sec, collectSectionRows(sec, val)); err != nil {
			return err
		}
	}
	return nil
}

// MarshalFullExif returns the library JSON for ex, or nil when ex is nil.
func MarshalFullExif(ex *exif.Exif) (json.RawMessage, error) {
	if ex == nil {
		return nil, nil
	}
	return json.Marshal(ex)
}
