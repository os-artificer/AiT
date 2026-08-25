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

// Package models declares the data row for `ai model list` and the
// receiver-free formatting helpers used to render it. The Client.ListModels
// implementation stays in the parent ai package because it depends on the
// unexported endpoint type and the Client facade.
package models

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

// Row is one configured provider row for ListModels output.
type Row struct {
	Provider  provider.ID
	Model     string
	APIStyle  string
	Role      string
	Key       string
	Enabled   string
	Reachable string
}

// ProviderRole classifies id as one of: disabled, local, default, fallback, or "—".
func ProviderRole(cfg *config.Config, id provider.ID, hasKey bool, enabled bool) string {
	if !enabled {
		return RoleDisabled
	}
	if id == cfg.LocalProviderID() && cfg.TieredEnabled() {
		return RoleLocal
	}
	if id == cfg.DefaultProviderID() {
		return RoleDefault
	}
	for _, fb := range cfg.AI.Fallback {
		if fb == id {
			return RoleFallback
		}
	}
	if hasKey {
		return RoleNeutral
	}
	return RoleNeutral
}

// KeyLabel returns the "yes"/"no"/"n/a" label for the KEY column.
// Ollama always shows "n/a" because it does not consume an API key.
func KeyLabel(id provider.ID, hasKey bool) string {
	if id == provider.Ollama {
		return KeyNA
	}
	if hasKey {
		return KeyYes
	}
	return KeyNo
}

// FormatTable renders rows as the human-readable `aissh ai model list` table.
func FormatTable(rows []Row) string {
	var b strings.Builder
	b.WriteString(TableHeader)
	for _, r := range rows {
		fmt.Fprintf(&b, "  %-9s  %-15s  %-4s  %-7s  %-9s  %s\n",
			r.Provider, TruncateDisplay(r.Model, ModelColumnWidth), r.Key, r.Enabled, r.Role, r.Reachable)
	}
	return b.String()
}

// TruncateDisplay shortens s to fit width max, appending "..." when truncated.
func TruncateDisplay(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-len(Ellipsis)] + Ellipsis
}
