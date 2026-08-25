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

package ollama

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// StatusReport holds Ollama configuration and runtime status for PrintStatus.
type StatusReport struct {
	TieredEnabled   bool
	ProviderEnabled bool
	BaseURL         string
	Model           string
	ModelsDir       string
	BinaryOK        bool
	APIOK           bool
	Detail          string
	PulledModels    []string
}

// BuildStatus probes config and the Ollama API without mutating config or starting serve.
func BuildStatus(ctx context.Context, cfg *config.Config, client *http.Client) StatusReport {
	if client == nil {
		client = &http.Client{Timeout: httpClientTimeout}
	}
	baseURL := BaseURLFromConfig(cfg)
	model := ModelFromConfig(cfg)
	avail := CheckAvailable(ctx, client, baseURL)

	report := StatusReport{
		TieredEnabled:   cfg != nil && cfg.TieredEnabled(),
		ProviderEnabled: cfg == nil || cfg.ProviderEnabled(config.ProviderOllama),
		BaseURL:         baseURL,
		Model:           model,
		ModelsDir:       ModelsDirFromConfig(cfg),
		BinaryOK:        avail.BinaryOK,
		APIOK:           avail.APIOK,
		Detail:          avail.Detail,
	}
	if avail.APIOK {
		tags, note, err := FetchTags(ctx, client, baseURL, model)
		if err == nil && note == "" {
			report.PulledModels = tags
		}
	} else if avail.Detail != "" {
		report.Detail = avail.Detail
	} else {
		report.Detail = unreachableNote
	}
	return report
}

// PrintStatus writes a human-readable status report to w.
func PrintStatus(ctx context.Context, cfg *config.Config, w io.Writer) error {
	report := BuildStatus(ctx, cfg, nil)
	return formatStatus(w, report)
}

func formatStatus(w io.Writer, r StatusReport) error {
	yesNo := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	lines := []string{
		"Ollama status:",
		fmt.Sprintf("  ai.tiered.enabled:            %s", yesNo(r.TieredEnabled)),
		fmt.Sprintf("  ai.providers.ollama.enabled:  %s", yesNo(r.ProviderEnabled)),
		fmt.Sprintf("  base_url:                     %s", r.BaseURL),
		fmt.Sprintf("  model:                        %s", r.Model),
		fmt.Sprintf("  models_dir (OLLAMA_MODELS):   %s", modelsDirLine(r.ModelsDir)),
		fmt.Sprintf("  binary:                       %s", yesNo(r.BinaryOK)),
		fmt.Sprintf("  api:                          %s", yesNo(r.APIOK)),
	}
	if r.Detail != "" && !r.APIOK {
		lines = append(lines, fmt.Sprintf("  detail:                       %s", r.Detail))
	}
	if len(r.PulledModels) > 0 {
		lines = append(lines, "  pulled_models:")
		for _, t := range r.PulledModels {
			lines = append(lines, "    "+t)
		}
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

func modelsDirLine(dir string) string {
	if dir == "" {
		return "(not set — run: " + toolname.AICLI("ollama setup") + ")"
	}
	return dir
}
