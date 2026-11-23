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
		{"host snapshot", Read},
		{"host cpu", Read},
		{"host proc --limit 5", Read},
		{"host top", Interactive},
		{"log search error", Read},
		{"log sources", Read},
		{"log -f search x", Interactive},
		{"log --follow search x", Interactive},
		{"img /tmp/a.jpg", Read},
		{"vid frames ./clip.mp4", Read},
		{"vid deps", Read},
		{"vid deps --install", Low},
		{"aitvid deps --install", Low},
		{"vid search ./c.mp4 --query x", Low},
		{"vid search ./c.mp4 --query x --no-save-hits", Read},
		{"aitvid search ./c.mp4 --query x --no-save-hits", Read},
		{"ssh session --list", Read},
		{"ssh session --delete x", High},
		{"ssh get user@h:/a ./a", Low},
		{"aihost snapshot", Read},
		{"aissh session --list", Read},
		{"unknownbin foo", Low},
	}
	for _, tc := range cases {
		got := ClassifyToolbox(tc.line)
		if got != tc.want {
			t.Errorf("ClassifyToolbox(%q)=%v want %v", tc.line, got, tc.want)
		}
	}
}
