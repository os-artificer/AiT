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
	"strings"
	"testing"
)

func TestAppendValueRows_StructuredUserComment(t *testing.T) {
	payload := `{"AIGC":{"Label":"1"}}`
	enc, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var rows []tableRow
	appendValueRows(&rows, "UserComment", enc, 0)
	if len(rows) != 1 || rows[0].field != "UserComment.AIGC.Label" || rows[0].value != "1" {
		t.Fatalf("rows=%v", rows)
	}
}

func TestAppendValueRows_RawUserComment(t *testing.T) {
	enc, err := json.Marshal("plain note")
	if err != nil {
		t.Fatal(err)
	}
	var rows []tableRow
	appendValueRows(&rows, "UserComment", enc, 0)
	if len(rows) != 1 || rows[0].field != "UserComment" || rows[0].value != "plain note" {
		t.Fatalf("rows=%v", rows)
	}
}

func TestAppendValueRows_KeepsFullLongScalar(t *testing.T) {
	long := strings.Repeat("x", maxValueLineRunes+10)
	enc, err := json.Marshal(long)
	if err != nil {
		t.Fatal(err)
	}
	var rows []tableRow
	appendValueRows(&rows, "Signature", enc, 0)
	if len(rows) != 1 || rows[0].value != long {
		t.Fatalf("want full value, got len=%d rows=%v", len(rows[0].value), rows)
	}
	if strings.Contains(rows[0].value, "…") {
		t.Fatal("must not truncate")
	}
}

func TestWriteTable_WrapsLongValueFully(t *testing.T) {
	long := strings.Repeat("a", maxValueLineRunes*2+5)
	var buf strings.Builder
	if err := writeTable(&buf, []tableRow{{field: "Signature", value: long}}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "…") {
		t.Fatal("must not truncate")
	}

	var collected strings.Builder
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " "), "Signature") {
			idx := strings.Index(line, tableColGap)
			if idx < 0 {
				t.Fatalf("bad row: %q", line)
			}
			collected.WriteString(line[idx+len(tableColGap):])
			continue
		}
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" || trimmed == tableFieldHeader+tableColGap+tableValueHeader || strings.HasPrefix(trimmed, "-") {
			continue
		}
		if strings.HasPrefix(line, " ") {
			collected.WriteString(trimmed)
		}
	}
	if collected.String() != long {
		t.Fatalf("wrapped output lost content: got %d want %d\n%s", len(collected.String()), len(long), out)
	}
	if strings.Count(out, "\n") < 3 {
		t.Fatalf("expected wrapped newlines, got %q", out)
	}
}
