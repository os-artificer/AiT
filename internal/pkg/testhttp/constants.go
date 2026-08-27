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

package testhttp

const (
	// ScenarioOK serves successful OpenAI, Ollama, and Nominatim payloads.
	ScenarioOK = "ok"
	// ScenarioUnauthorized returns HTTP 401 on chat completions.
	ScenarioUnauthorized = "unauthorized"
	// ScenarioRateLimited returns HTTP 429 on chat completions.
	ScenarioRateLimited = "rate_limited"
	// ScenarioServerError returns HTTP 500 on every route.
	ScenarioServerError = "server_error"
	// ScenarioNotFound returns HTTP 404 on every route.
	ScenarioNotFound = "not_found"
	// ScenarioMalformed returns HTTP 200 with invalid JSON.
	ScenarioMalformed = "malformed"
	// ScenarioGeocodeUnable returns Nominatim {"error":"Unable to geocode"}.
	ScenarioGeocodeUnable = "geocode_unable"
	// ScenarioHang blocks until the request context is cancelled.
	ScenarioHang = "hang"

	defaultChatReply    = "mock-chat-ok"
	defaultDisplayName  = "上海市黄浦区外滩, 上海, 中国"
	defaultOllamaModel  = "llama3.2:latest"
	pathChatCompletions = "/v1/chat/completions"
	pathAPITags         = "/api/tags"
	pathReverse         = "/reverse"
)
