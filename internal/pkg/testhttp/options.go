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

// Config selects which external-HTTP scenario the mock serves.
type Config struct {
	// Addr is the listen address for Listen; empty uses 127.0.0.1:0.
	Addr string
	// Scenario is one of the Scenario* constants; empty uses ScenarioOK.
	Scenario string
	// ChatReply is the OpenAI assistant content for ScenarioOK.
	ChatReply string
	// DisplayName is the Nominatim display_name for ScenarioOK.
	DisplayName string
	// OllamaModel is the /api/tags model name for ScenarioOK.
	OllamaModel string
}

// Info is the JSON payload printed by cmd/testenv-mock.
type Info struct {
	BaseURL     string `json:"base_url"`
	ChatURL     string `json:"chat_url"`
	TagsURL     string `json:"tags_url"`
	ReverseURL  string `json:"reverse_url"`
	Scenario    string `json:"scenario"`
	ChatReply   string `json:"chat_reply"`
	DisplayName string `json:"display_name"`
}
