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
	"encoding/json"
	"io"
	"strings"
)

// DefaultSections returns CPU, Mem, Disk, Net, Proto, Proc sections in display order.
func DefaultSections() []Section {
	return []Section{
		CPUSection{},
		MemSection{},
		DiskSection{},
		NetSection{},
		ProtoSection{},
		ProcSection{},
	}
}

// SectionByID returns the section for id, or nil if unknown.
func SectionByID(id string) Section {
	for _, s := range DefaultSections() {
		if s.ID() == id {
			return s
		}
	}
	return nil
}

// WriteHuman writes selected sections (or all default) as human text to w.
// Returns an error if a write fails.
func WriteHuman(w io.Writer, snap Snapshot, sections []Section) error {
	if len(sections) == 0 {
		sections = DefaultSections()
		if err := writeHostHeaderLine(w, snap); err != nil {
			return err
		}
	}
	for i, sec := range sections {
		if err := writeHumanSection(w, sec, snap, i > 0); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON encodes snap or a single section value as indented JSON to w.
// When section is nil, the full Snapshot is encoded.
func WriteJSON(w io.Writer, snap Snapshot, section Section) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if section == nil {
		return enc.Encode(snap)
	}
	wrapper := map[string]any{
		"time":       snap.Time,
		section.ID(): section.JSONValue(snap),
	}
	return enc.Encode(wrapper)
}

func writeHostHeaderLine(w io.Writer, snap Snapshot) error {
	line := FormatHostHeader(snap.Host, snap.Mem.Total, snap.Time)
	if line == "" {
		return nil
	}
	_, err := io.WriteString(w, line+"\n")
	return err
}

func writeHumanSection(w io.Writer, sec Section, snap Snapshot, leadNL bool) error {
	if leadNL {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	text := sec.Human(snap)
	if _, err := io.WriteString(w, text); err != nil {
		return err
	}
	if strings.HasSuffix(text, "\n") {
		return nil
	}
	_, err := io.WriteString(w, "\n")
	return err
}
