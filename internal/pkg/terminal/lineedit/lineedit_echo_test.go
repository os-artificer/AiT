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

package lineedit

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal/screen"
)

func TestAppendRedrawLine_Golden(t *testing.T) {
	got := string(AppendRedrawLine(nil, "AiT> ", []rune("ab"), 1))
	want := "\rAiT> ab\x1b[K\b"
	if got != want {
		t.Fatalf("AppendRedrawLine = %q want %q", got, want)
	}
}

func TestAppendClearAndRedrawLine_Golden(t *testing.T) {
	got := string(AppendClearAndRedrawLine(nil, "p> ", []rune("x"), 1))
	if !strings.HasPrefix(got, screen.ClearSeq()) {
		t.Fatalf("missing clear seq in %q", got)
	}
	if !strings.Contains(got, "\rp> x") {
		t.Fatalf("missing redraw in %q", got)
	}
}
