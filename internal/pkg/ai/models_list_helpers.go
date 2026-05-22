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

package ai

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func providerRole(cfg *config.Config, id ProviderID, hasKey bool, enabled bool) string {
	if !enabled {
		return "disabled"
	}
	if id == cfg.LocalProviderID() && cfg.TieredEnabled() {
		return "local"
	}
	if id == cfg.DefaultProviderID() {
		return "default"
	}
	for _, fb := range cfg.AI.Fallback {
		if fb == id {
			return "fallback"
		}
	}
	if hasKey {
		return "—"
	}
	return "—"
}

func keyLabel(id ProviderID, hasKey bool) string {
	if id == ProviderOllama {
		return "n/a"
	}
	if hasKey {
		return "yes"
	}
	return "no"
}

func formatModelsTable(rows []ModelRow) string {
	const hdr = "AiT configured models:\n  PROVIDER   MODEL           KEY   ENABLED  ROLE       REACHABLE\n"
	var b strings.Builder
	b.WriteString(hdr)
	for _, r := range rows {
		fmt.Fprintf(&b, "  %-9s  %-15s  %-4s  %-7s  %-9s  %s\n",
			r.Provider, truncateDisplay(r.Model, 15), r.Key, r.Enabled, r.Role, r.Reachable)
	}
	return b.String()
}

func truncateDisplay(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
