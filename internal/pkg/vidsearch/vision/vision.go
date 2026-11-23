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

// Package vision analyzes video frames with a vision-capable LLM.
package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

// ErrNotVisionModel is returned when the configured model cannot accept images.
var ErrNotVisionModel = errhint.NewSentinel("model is not vision-capable")

// Analyzer runs object/text vision search over frames.
type Analyzer interface {
	Analyze(ctx context.Context, req Request) ([]hit.Hit, error)
}

// Request is a batch vision analysis request.
type Request struct {
	Query  string
	Frames []frames.Frame
	Mode   string
}

// Config for OpenAI-compatible vision chat.
type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
	// ForceVision skips the soft capability check (tests).
	ForceVision bool
}

// Client implements Analyzer via OpenAI-compatible chat completions.
type Client struct {
	cfg Config
}

// NewClient returns a vision Analyzer. Model must be non-empty.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errhint.WithTry(ErrNotVisionModel,
			"Set ai.vision.model to a vision-capable model id.",
			"ait agent config set ai.vision.model <model>",
		)
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	return &Client{cfg: cfg}, nil
}

// Analyze asks the model whether each batch of frames matches the query.
func (c *Client) Analyze(ctx context.Context, req Request) ([]hit.Hit, error) {
	if !c.cfg.ForceVision && looksTextOnlyModel(c.cfg.Model) {
		return nil, errhint.WithTry(ErrNotVisionModel,
			fmt.Sprintf("Model %q does not look vision-capable.", c.cfg.Model),
			"ait agent config set ai.vision.model <vision-model>",
			"aitvid search … --mode text --local-only",
		)
	}
	const batch = 4
	var hits []hit.Hit
	for i := 0; i < len(req.Frames); i += batch {
		end := i + batch
		if end > len(req.Frames) {
			end = len(req.Frames)
		}
		chunk := req.Frames[i:end]
		partHits, err := c.analyzeBatch(ctx, req.Query, req.Mode, chunk)
		if err != nil {
			return hits, err
		}
		hits = append(hits, partHits...)
	}
	return hits, nil
}

func looksTextOnlyModel(model string) bool {
	m := strings.ToLower(model)
	if strings.Contains(m, "vision") || strings.Contains(m, "gpt-4o") || strings.Contains(m, "llava") {
		return false
	}
	// Conservative: deepseek-chat / llama3.2 without vision tag → text-only.
	if strings.Contains(m, "deepseek-chat") || strings.Contains(m, "minimax") {
		return true
	}
	return false
}

func (c *Client) analyzeBatch(ctx context.Context, query, mode string, fr []frames.Frame) ([]hit.Hit, error) {
	content := []map[string]any{
		{"type": "text", "text": fmt.Sprintf(
			"You search video frames. Query: %q. For each image index (0-based) say HIT or MISS and a short English evidence. Reply JSON array: [{\"i\":0,\"hit\":true,\"snippet\":\"...\"}]",
			query)},
	}
	for _, f := range fr {
		b64, err := fileBase64(f.Path)
		if err != nil {
			continue
		}
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": "data:image/jpeg;base64," + b64,
			},
		})
	}
	body := map[string]any{
		"model": c.cfg.Model,
		"messages": []map[string]any{
			{"role": "user", "content": content},
		},
	}
	raw, _ := json.Marshal(body)
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vision API HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Choices) == 0 {
		return nil, nil
	}
	return parseHitJSON(parsed.Choices[0].Message.Content, fr, query, mode)
}

type hitRow struct {
	I       int    `json:"i"`
	Hit     bool   `json:"hit"`
	Snippet string `json:"snippet"`
}

func parseHitJSON(content string, fr []frames.Frame, query, mode string) ([]hit.Hit, error) {
	content = strings.TrimSpace(content)
	if i := strings.Index(content, "["); i >= 0 {
		if j := strings.LastIndex(content, "]"); j > i {
			content = content[i : j+1]
		}
	}
	var rows []hitRow
	if err := json.Unmarshal([]byte(content), &rows); err != nil {
		return nil, nil // soft: no hits if unparseable
	}
	var hits []hit.Hit
	for _, r := range rows {
		if !r.Hit || r.I < 0 || r.I >= len(fr) {
			continue
		}
		hits = append(hits, hit.Hit{
			TSec:    fr[r.I].TSec,
			Frame:   fr[r.I].Path,
			Kind:    hit.KindObject,
			Query:   query,
			Snippet: r.Snippet,
			Score:   0.8,
			Mode:    mode,
		})
	}
	return hits, nil
}

func fileBase64(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
