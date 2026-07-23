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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	tableFieldHeader         = "Field"
	tableValueHeader         = "Value"
	maxStructuredExpandDepth = 3
	tableColGap              = "  "
	tableValueSepWidth       = 5
	// maxValueLineRunes wraps long values across lines; content is never truncated.
	maxValueLineRunes = 72
	emptyTimeJSON     = "0001-01-01T00:00:00Z"
	fieldPathSep      = "."
)

type tableRow struct {
	field string
	value string
}

// writeNamedTable writes an optional section title, a Field/Value table, and a trailing blank line.
// name empty skips the title (root keys). Empty rows are a no-op.
func writeNamedTable(w io.Writer, name string, rows []tableRow) error {
	if len(rows) == 0 {
		return nil
	}
	if name != "" {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	if err := writeTable(w, rows); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}

func writeTable(w io.Writer, rows []tableRow) error {
	fieldW := len(tableFieldHeader)
	for _, r := range rows {
		if n := len(r.field); n > fieldW {
			fieldW = n
		}
	}

	if _, err := fmt.Fprintf(w, "%-*s%s%s\n", fieldW, tableFieldHeader, tableColGap, tableValueHeader); err != nil {
		return err
	}
	sep := strings.Repeat("-", fieldW) + tableColGap + strings.Repeat("-", tableValueSepWidth)
	if _, err := fmt.Fprintln(w, sep); err != nil {
		return err
	}

	contPad := strings.Repeat(" ", fieldW) + tableColGap
	for _, r := range rows {
		if err := writeTableRow(w, fieldW, contPad, r); err != nil {
			return err
		}
	}
	return nil
}

func writeTableRow(w io.Writer, fieldW int, contPad string, r tableRow) error {
	lines := wrapValueLines(r.value, maxValueLineRunes)
	if _, err := fmt.Fprintf(w, "%-*s%s%s\n", fieldW, r.field, tableColGap, lines[0]); err != nil {
		return err
	}
	for _, line := range lines[1:] {
		if _, err := fmt.Fprintf(w, "%s%s\n", contPad, line); err != nil {
			return err
		}
	}
	return nil
}

func collectRootRows(top map[string]json.RawMessage) []tableRow {
	var rows []tableRow
	for _, key := range fullRootKeys {
		val, ok := top[key]
		if !ok || jsonValueEmpty(val) {
			continue
		}
		appendValueRows(&rows, key, val, 0)
	}
	return rows
}

func collectSectionRows(name string, raw json.RawMessage) []tableRow {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		var rows []tableRow
		appendValueRows(&rows, name, raw, 0)
		return rows
	}

	keys := nonEmptyObjectKeys(obj)
	if len(keys) == 0 {
		return nil
	}

	var rows []tableRow
	for _, k := range keys {
		appendValueRows(&rows, k, obj[k], 0)
	}
	return rows
}

func appendValueRows(rows *[]tableRow, prefix string, raw json.RawMessage, depth int) {
	raw = bytes.TrimSpace(raw)
	if jsonValueEmpty(raw) {
		return
	}

	if appendIfJSONObject(rows, prefix, raw, depth) {
		return
	}
	if appendIfJSONArray(rows, prefix, raw, depth) {
		return
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if depth < maxStructuredExpandDepth {
			if expanded, ok := tryExpandStructuredString(prefix, s, depth+1); ok {
				*rows = append(*rows, expanded...)
				return
			}
		}
		*rows = append(*rows, tableRow{field: prefix, value: s})
		return
	}

	*rows = append(*rows, tableRow{
		field: prefix,
		value: formatJSONScalar(raw),
	})
}

func appendIfJSONObject(rows *[]tableRow, prefix string, raw json.RawMessage, depth int) bool {
	if len(raw) == 0 || raw[0] != '{' {
		return false
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}
	for _, k := range nonEmptyObjectKeys(obj) {
		appendValueRows(rows, joinFieldPath(prefix, k), obj[k], depth)
	}
	return true
}

func appendIfJSONArray(rows *[]tableRow, prefix string, raw json.RawMessage, depth int) bool {
	if len(raw) == 0 || raw[0] != '[' {
		return false
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return false
	}
	for i, el := range arr {
		if jsonValueEmpty(el) {
			continue
		}
		appendValueRows(rows, fmt.Sprintf("%s[%d]", prefix, i), el, depth)
	}
	return true
}

func joinFieldPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + fieldPathSep + key
}

// wrapValueLines splits s on newlines then wraps each segment to at most maxRunes.
// All runes are preserved; nothing is truncated.
func wrapValueLines(s string, maxRunes int) []string {
	if s == "" {
		return []string{""}
	}
	if maxRunes <= 0 {
		return strings.Split(s, "\n")
	}

	var out []string
	for _, para := range strings.Split(s, "\n") {
		out = append(out, wrapRunes(para, maxRunes)...)
	}
	return out
}

func wrapRunes(s string, maxRunes int) []string {
	if s == "" {
		return []string{""}
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return []string{s}
	}
	out := make([]string, 0, (len(runes)+maxRunes-1)/maxRunes)
	for len(runes) > 0 {
		n := maxRunes
		if n > len(runes) {
			n = len(runes)
		}
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}

func nonEmptyObjectKeys(obj map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(obj))
	for k, v := range obj {
		if jsonValueEmpty(v) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatJSONScalar(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

func jsonValueEmpty(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return true
	}
	if bytes.Equal(raw, []byte(`""`)) || bytes.Equal(raw, []byte("[]")) || bytes.Equal(raw, []byte("{}")) {
		return true
	}
	if bytes.Equal(raw, []byte("0")) || bytes.Equal(raw, []byte("0.0")) || bytes.Equal(raw, []byte("false")) {
		return true
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s == "" || s == emptyTimeJSON
	}

	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return f == 0
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		return jsonArrayEmpty(arr)
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		return jsonObjectEmpty(obj)
	}
	return false
}

func jsonArrayEmpty(arr []json.RawMessage) bool {
	if len(arr) == 0 {
		return true
	}
	for _, el := range arr {
		if !jsonValueEmpty(el) {
			return false
		}
	}
	return true
}

func jsonObjectEmpty(obj map[string]json.RawMessage) bool {
	for _, v := range obj {
		if !jsonValueEmpty(v) {
			return false
		}
	}
	return true
}
