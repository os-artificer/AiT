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

import "time"

// Binary and install — executable name and download URLs.
const (
	binaryName         = "ollama"
	installScriptURL   = "https://ollama.com/install.sh"
	windowsDownloadURL = "https://ollama.com/download/windows"
)

// Timeouts — probe, pull, and install waits.
const (
	httpClientTimeout = 3 * time.Second
	probeInterval     = 2 * time.Second
	probeTotalTimeout = 30 * time.Second
	pullTimeout       = 30 * time.Minute
	installTimeout    = 15 * time.Minute
	serveStartWait    = 5 * time.Second
)

// API — Ollama HTTP paths and status messages.
const (
	pathAPITags         = "/api/tags"
	unreachableNote     = "ollama: unreachable (is ollama serve running?)"
	logStepAPIReachable = "API: reachable"
)
