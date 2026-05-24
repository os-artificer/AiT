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

package errhint

import (
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

func TestWithTry_PreservesIs(t *testing.T) {
	err := WithTry(store.ErrSessionNotFound, "The list ID column is for display only.",
		"ait session --load <alias>",
		"ait session --help",
	)
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
	s := err.Error()
	if !strings.Contains(s, "\n\n") {
		t.Fatalf("expected blank line separation: %q", s)
	}
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, "display only") {
		t.Fatalf("missing explanation: %q", s)
	}
}

func TestFormatTryBlock_Empty(t *testing.T) {
	if got := FormatTryBlock(); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestFormatTryBlock_Lines(t *testing.T) {
	got := FormatTryBlock("ait session --list", "ait session --help")
	if !strings.Contains(got, "Try:\n  ait session --list") {
		t.Fatalf("got %q", got)
	}
}
