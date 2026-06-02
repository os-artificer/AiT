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
	"net/http"
	"runtime"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// Availability reports Ollama binary and API reachability.
type Availability struct {
	BinaryOK bool
	APIOK    bool
	Detail   string
}

// CheckAvailable probes PATH and GET /api/tags.
func CheckAvailable(ctx context.Context, client *http.Client, baseURL string) Availability {
	out := Availability{}
	if _, err := platform.LookPath(binaryName); err == nil {
		out.BinaryOK = true
	}
	if client == nil {
		client = &http.Client{Timeout: httpClientTimeout}
	}
	_, note, err := FetchTags(ctx, client, baseURL, "")
	if err == nil && note == "" {
		out.APIOK = true
		return out
	}
	if !out.BinaryOK {
		out.Detail = "ollama binary not in PATH"
	} else {
		out.Detail = "ollama API not reachable"
	}
	return out
}

// WaitReady polls /api/tags until OK or timeout; may start ollama serve on Unix.
func WaitReady(ctx context.Context, client *http.Client, baseURL, modelsDir string) error {
	if client == nil {
		client = &http.Client{Timeout: httpClientTimeout}
	}
	deadline := time.Now().Add(probeTotalTimeout)
	for time.Now().Before(deadline) {
		_, note, err := FetchTags(ctx, client, baseURL, "")
		if err == nil && note == "" {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(probeInterval):
		}
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if _, err := platform.LookPath(binaryName); err == nil {
			_ = StartServeDetached(modelsDir)
			waitCtx, cancel := context.WithTimeout(ctx, serveStartWait)
			defer cancel()
			<-waitCtx.Done()
			_, note, err := FetchTags(ctx, client, baseURL, "")
			if err == nil && note == "" {
				return nil
			}
		}
	}

	return fmt.Errorf("ollama API not reachable at %s (start Ollama app or run: ollama serve)", TagsBaseURL(baseURL))
}
