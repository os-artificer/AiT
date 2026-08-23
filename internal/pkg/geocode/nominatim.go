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

// Package geocode provides reverse geocoding via OpenStreetMap Nominatim.
package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://nominatim.openstreetmap.org/reverse"
	acceptLanguage   = "zh-CN,en"
	httpTimeout      = 8 * time.Second
	maxResponseBytes = 1 << 20
	projectHomeURL   = "https://github.com/os-artificer/AiT"
	defaultUserAgent = "AiT-aiimg/dev (" + projectHomeURL + ")"
)

// Client performs Nominatim reverse geocode requests.
type Client struct {
	// BaseURL overrides the Nominatim reverse endpoint (tests / mirrors).
	BaseURL string
	// UserAgent is sent on each request; empty uses defaultUserAgent.
	UserAgent string
	// HTTP is the optional client; nil uses a client with httpTimeout.
	HTTP *http.Client
}

// Reverse looks up a human-readable address for lat/lon.
// Returns an error when the HTTP call fails or display_name is empty.
func (c *Client) Reverse(ctx context.Context, lat, lon float64) (string, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}

	ua := c.UserAgent
	if ua == "" {
		ua = defaultUserAgent
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: httpTimeout}
	}

	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("format", "jsonv2")
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept-Language", acceptLanguage)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nominatim: HTTP %d", resp.StatusCode)
	}

	var parsed struct {
		DisplayName string `json:"display_name"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("nominatim: %s", parsed.Error)
	}

	name := strings.TrimSpace(parsed.DisplayName)
	if name == "" {
		return "", fmt.Errorf("nominatim: empty display_name")
	}
	return name, nil
}
