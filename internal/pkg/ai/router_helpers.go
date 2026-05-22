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
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"

	openai "github.com/sashabaranov/go-openai"
)

const envFailoverLog = "AIT_AI_FAILOVER_LOG"

// retryable: network/timeout, HTTP 5xx/429, empty choices. Other API errors still try next provider.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 429, 500, 502, 503, 504:
			return true
		default:
			return apiErr.HTTPStatusCode >= 500
		}
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "empty response") || strings.Contains(msg, "empty choices") {
		return true
	}
	return false
}

func resolveTryOrder(cfg *config.Config, sessionOverride ProviderID, keyed map[ProviderID]struct{}) []ProviderID {
	primary := cfg.DefaultProviderID()
	if sessionOverride != "" {
		primary = sessionOverride
	}

	fallback := cfg.AI.Fallback
	if len(fallback) == 0 {
		fallback = config.CloudProviderIDs()
	}

	seen := make(map[ProviderID]bool)
	var order []ProviderID
	add := func(id ProviderID) {
		if id == "" || seen[id] {
			return
		}
		if _, ok := keyed[id]; !ok {
			return
		}
		seen[id] = true
		order = append(order, id)
	}

	add(primary)
	for _, id := range fallback {
		if id != primary {
			add(id)
		}
	}
	return order
}

func aggregateErrors(provider ProviderID, errs []error) error {
	if len(errs) == 0 {
		return fmt.Errorf("ai: all providers failed (last: %s)", provider)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ai: all providers failed; last attempt %s:", provider)
	for _, e := range errs {
		if e != nil {
			b.WriteString("\n  - ")
			b.WriteString(e.Error())
		}
	}
	return errors.New(b.String())
}

func logFailover(w io.Writer, from, to ProviderID, err error) {
	if os.Getenv(envFailoverLog) != "1" {
		return
	}
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, "AiT: AI failover %s -> %s (%v)\n", from, to, err)
}

func apiKeyHint(provider ProviderID) string {
	return fmt.Sprintf("ait config set ai.%s.api_key <key>", provider)
}
