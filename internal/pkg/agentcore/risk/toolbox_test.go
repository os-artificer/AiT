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

package risk

import "testing"

func TestClassify_ToolboxRead(t *testing.T) {
	cases := []struct {
		line string
		want Level
	}{
		{"aihost snapshot", Read},
		{"aihost cpu", Read},
		{"aihost proc --limit 5", Read},
		{"aihost top", Interactive},
		{"ailog search error", Read},
		{"ailog sources", Read},
		{"ailog -f search x", Interactive},
		{"ailog --follow search x", Interactive},
		{"aiimg /tmp/a.jpg", Read},
		{"aissh session --list", Read},
		{"aissh session --delete x", High},
		{"aissh get user@h:/a ./a", Low},
		{"unknownbin foo", Low},
	}
	for _, tc := range cases {
		got := ClassifyToolbox(tc.line)
		if got != tc.want {
			t.Errorf("ClassifyToolbox(%q)=%v want %v", tc.line, got, tc.want)
		}
	}
}
