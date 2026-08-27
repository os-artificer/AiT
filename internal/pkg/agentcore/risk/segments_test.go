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

import (
	"strings"
	"testing"
)

// AG-10: SplitSegments splits on unquoted operators only.
func TestSplitSegments_QuotesAndOperators(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cmd  string
		want []string
	}{
		{"ls -la", []string{"ls -la"}},
		{"ls; rm -rf /", []string{"ls", "rm -rf /"}},
		{"git log && git push", []string{"git log", "git push"}},
		{"cat a || cat b", []string{"cat a", "cat b"}},
		{"ps aux | grep go", []string{"ps aux", "grep go"}},
		{"true ;; false", []string{"true", "false"}},
		{"echo 'a; b' ", []string{"echo 'a; b'"}},
		{`echo "x && y"`, []string{`echo "x && y"`}},
		{`echo a\; b`, []string{`echo a\; b`}},
		{"ls\nrm -rf /", []string{"ls", "rm -rf /"}},
	}

	for _, tc := range cases {
		got := SplitSegments(tc.cmd)
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Fatalf("SplitSegments(%q) = %q, want %q", tc.cmd, got, tc.want)
		}
	}
}

// AG-10: blank input yields no segments.
func TestSplitSegments_Blank(t *testing.T) {
	t.Parallel()
	if got := SplitSegments("   "); got != nil {
		t.Fatalf("expected nil for blank input, got %q", got)
	}
}

// AG-10: pathological operator runs stay within maxSegments.
func TestSplitSegments_Capped(t *testing.T) {
	t.Parallel()
	cmd := strings.TrimSuffix(strings.Repeat("ls;", 4*maxSegments), ";")
	if got := len(SplitSegments(cmd)); got > maxSegments {
		t.Fatalf("got %d segments, want <= %d", got, maxSegments)
	}
}

// AG-8: only writing metacharacters disqualify the read-only fast path.
func TestHasUnsafeReadMeta(t *testing.T) {
	t.Parallel()
	unsafe := []string{
		"cat /etc/hosts > /tmp/out",
		"cat file >> /tmp/out",
		"echo $(rm -rf /tmp/x)",
		"echo `rm -rf /tmp/x`",
		"ping example.com &",
		"cat < /tmp/in",
	}
	for _, cmd := range unsafe {
		if !hasUnsafeReadMeta(cmd) {
			t.Fatalf("expected unsafe meta for %q", cmd)
		}
	}

	safe := []string{
		"ls -la",
		"grep pattern file 2>&1",
		"grep pattern file > /dev/null",
		"echo 'a > b'",
		`echo "x & y"`,
	}
	for _, cmd := range safe {
		if hasUnsafeReadMeta(cmd) {
			t.Fatalf("expected safe for %q", cmd)
		}
	}
}
