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

package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/ai/models"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/ollama"
)

// ModelRow is one configured provider row for ListModels output (alias of models.Row).
type ModelRow = models.Row

// ModelsReport is the structured result of ListModels.
type ModelsReport struct {
	Rows       []ModelRow
	OllamaTags []string
	TagsNote   string
}

// ListModels builds a report of providers AiT can use and optional Ollama tag names.
func (c *Client) ListModels(ctx context.Context) (*ModelsReport, error) {
	if c == nil || c.cfg == nil {
		return nil, fmt.Errorf("%s", errClientNotConfigured)
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: modelsListHTTPTimeout}
	}

	var rows []ModelRow
	for _, id := range config.AllProviderIDs() {
		pc, inCfg := providerConfigForList(c.cfg, id)
		if !inCfg {
			continue
		}

		hasKey := id == ProviderOllama
		if !hasKey {
			_, err := creds.GetProviderAPIKey(id)
			hasKey = err == nil
		}
		if id == ProviderOllama {
			if !inCfg && !c.cfg.TieredEnabled() {
				continue
			}
		} else if !hasKey && !inCfg {
			continue
		}

		enabled := c.cfg.ProviderEnabled(id)
		enabledLabel := models.KeyYes
		if !enabled {
			enabledLabel = models.KeyNo
		}

		reachable := c.providerReachable(ctx, httpClient, id, enabled)

		style := pc.APIStyle
		if style == "" {
			style = config.APIStyleOpenAI
		}
		rows = append(rows, ModelRow{
			Provider:  id,
			Model:     pc.Model,
			APIStyle:  style,
			Role:      models.ProviderRole(c.cfg, id, hasKey, enabled),
			Key:       models.KeyLabel(id, hasKey),
			Enabled:   enabledLabel,
			Reachable: reachable,
		})
	}

	report := &ModelsReport{Rows: rows}
	if pc, ok := c.cfg.Provider(ProviderOllama); ok && c.cfg.TieredEnabled() && c.cfg.ProviderEnabled(ProviderOllama) {
		tags, note, _ := ollama.FetchTags(ctx, httpClient, pc.BaseURL, pc.Model)
		report.OllamaTags = tags
		report.TagsNote = note
	}
	return report, nil
}

// Format writes a human-readable models report.
func (r *ModelsReport) Format() string {
	if r == nil {
		return ""
	}
	out := models.FormatTable(r.Rows)
	if r.TagsNote != "" {
		out += "\n" + r.TagsNote + "\n"
	} else if len(r.OllamaTags) > 0 {
		out += "\n" + models.OllamaTagsHeader
		for _, t := range r.OllamaTags {
			out += "  " + t + "\n"
		}
	}
	return out
}
