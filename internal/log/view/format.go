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

package view

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

const (
	// humanTimeLayout is the timestamp format for the default output.
	humanTimeLayout = "2006-01-02 15:04:05"
	// unknownField stands in for an attribute the backend did not provide.
	unknownField = "-"
)

// jsonRecord is the --json shape. Fields the backend did not provide are
// omitted rather than emitted empty.
type jsonRecord struct {
	Time    string `json:"time,omitempty"`
	Host    string `json:"host,omitempty"`
	Source  string `json:"source,omitempty"`
	Unit    string `json:"unit,omitempty"`
	Level   string `json:"level,omitempty"`
	PID     int    `json:"pid,omitempty"`
	Message string `json:"message"`
	Raw     string `json:"raw,omitempty"`
}

// formatRecord renders rec as one output line, as JSON when asJSON is set.
func formatRecord(rec logsrc.Record, q logsrc.Query, asJSON bool) string {
	if asJSON {
		return formatJSON(rec, q)
	}
	return formatHuman(rec, q)
}

func formatHuman(rec logsrc.Record, q logsrc.Query) string {
	// One record must be one stdout line so -n N matches `wc -l` and the pager
	// never splits a single match across frames.
	return fmt.Sprintf("%s %s [%s] %s",
		orUnknown(formatTime(rec.Time, q.UTC, humanTimeLayout)),
		orUnknown(rec.Unit),
		orUnknown(levelName(rec.Level)),
		flattenMessageLine(rec.Message))
}

func formatJSON(rec logsrc.Record, q logsrc.Query) string {
	jr := jsonRecord{
		Time:    formatTime(rec.Time, q.UTC, time.RFC3339Nano),
		Host:    rec.Host,
		Source:  rec.Source,
		Unit:    rec.Unit,
		Level:   levelName(rec.Level),
		PID:     rec.PID,
		Message: rec.Message,
	}
	if q.Raw {
		jr.Raw = rec.Raw
	}
	b, err := json.Marshal(jr)
	if err != nil {
		return fmt.Sprintf(`{"message":%q}`, rec.Message)
	}
	return string(b)
}

// formatTime renders t in the requested location, or "" when unknown.
func formatTime(t time.Time, utc bool, layout string) string {
	if t.IsZero() {
		return ""
	}
	loc := time.Local
	if utc {
		loc = time.UTC
	}
	return t.In(loc).Format(layout)
}

// levelName returns the severity name, or "" when the backend did not report
// one, so that JSON omits it and human output shows a placeholder.
func levelName(lv logsrc.Level) string {
	if lv == logsrc.LevelUnset {
		return ""
	}
	return lv.String()
}

func orUnknown(s string) string {
	if s == "" {
		return unknownField
	}
	return s
}

// flattenMessageLine replaces embedded newlines so human output stays one line
// per record. JSON keeps the original Message (including newlines).
func flattenMessageLine(s string) string {
	if !strings.ContainsAny(s, "\r\n") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "\r", " ")
}
