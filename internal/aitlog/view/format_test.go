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
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

func TestFormatHuman_CollapsesNewlines(t *testing.T) {
	rec := logsrc.Record{
		Time:    time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Unit:    "kernel",
		Level:   logsrc.LevelInfo,
		Message: "tcp connected\nso_gencnt: 1\r\nerror: 0",
	}
	line := formatRecord(rec, logsrc.Query{UTC: true}, false)
	if strings.ContainsAny(line, "\r\n") {
		t.Fatalf("human line must be single-line, got %q", line)
	}
	if !strings.Contains(line, "tcp connected so_gencnt: 1 error: 0") {
		t.Fatalf("%q", line)
	}
}

func TestFormatJSON_KeepsNewlines(t *testing.T) {
	rec := logsrc.Record{Message: "a\nb"}
	line := formatRecord(rec, logsrc.Query{}, true)
	if !strings.Contains(line, `\n`) {
		t.Fatalf("json should escape newline in message, got %q", line)
	}
}
