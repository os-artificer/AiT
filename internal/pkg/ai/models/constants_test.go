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

package models

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
)

// TestConstants_TableHeaderGolden (C14) keeps model list table header stable.
func TestConstants_TableHeaderGolden(t *testing.T) {
	out := FormatTable([]Row{{
		Provider:  provider.Deepseek,
		Model:     "deepseek-chat",
		Key:       KeyYes,
		Enabled:   KeyYes,
		Role:      RoleDefault,
		Reachable: KeyYes,
	}})
	if !strings.HasPrefix(out, TableHeader) {
		t.Fatalf("missing table header: %q", out)
	}
	if !strings.Contains(out, "deepseek") {
		t.Fatal("missing provider row")
	}
}
