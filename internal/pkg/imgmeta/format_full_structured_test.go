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
	"strings"
	"testing"
)

func TestTryExpandStructuredString_AIGCJSON(t *testing.T) {
	payload := `{"AIGC":{"Label":"1","ProduceID":"voiceassistant-abc","ReservedCode1":"{\"SecurityData\":{\"Type\":\"TC260PG\"}}"}}`
	rows, ok := tryExpandStructuredString("UserComment", payload, 1)
	if !ok {
		t.Fatal("expected structured expand")
	}
	joined := flatFields(rows)
	for _, want := range []string{
		"UserComment.AIGC.Label",
		"UserComment.AIGC.ProduceID",
		"UserComment.AIGC.ReservedCode1.SecurityData.Type",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, rows)
		}
	}
	if strings.Contains(joined, `"AIGC"`) {
		t.Fatalf("should not keep raw JSON blob as single field: %v", rows)
	}
}

func TestTryExpandStructuredString_XML(t *testing.T) {
	payload := `<meta><AIGC Label="1"><ProduceID>voiceassistant-abc</ProduceID></AIGC></meta>`
	rows, ok := tryExpandStructuredString("UserComment", payload, 1)
	if !ok {
		t.Fatal("expected XML expand")
	}
	joined := flatFields(rows)
	for _, want := range []string{
		"UserComment.meta.AIGC.@Label",
		"UserComment.meta.AIGC.ProduceID",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, rows)
		}
	}
}

func TestTryExpandStructuredString_YAML(t *testing.T) {
	payload := "AIGC:\n  Label: \"1\"\n  ProduceID: voiceassistant-abc\n"
	rows, ok := tryExpandStructuredString("UserComment", payload, 1)
	if !ok {
		t.Fatal("expected YAML expand")
	}
	joined := flatFields(rows)
	for _, want := range []string{
		"UserComment.AIGC.Label",
		"UserComment.AIGC.ProduceID",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, rows)
		}
	}
}

func TestTryExpandStructuredString_RawFallback(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"plain", "hello from camera"},
		{"badJSON", "{not-json"},
		{"badXML", "<not-xml"},
		{"yamlBool", "true"},
		{"yamlInt", "42"},
		{"yamlString", "just-a-scalar"},
		{"emptyJSONObject", "{}"},
		{"emptyJSONArray", "[]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, ok := tryExpandStructuredString("UserComment", tc.payload, 1)
			if ok {
				t.Fatalf("expected raw fallback for %q, got rows=%v", tc.payload, rows)
			}
		})
	}
}

func flatFields(rows []tableRow) string {
	parts := make([]string, len(rows))
	for i, r := range rows {
		parts[i] = r.field + "=" + r.value
	}
	return strings.Join(parts, "\n")
}
