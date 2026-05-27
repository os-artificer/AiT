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

// Package provider declares the public AI provider identifiers and user-visible
// errors (sentinel + Try-block hints) shared by the multi-provider router,
// agent loop, and CLI surface.
package provider

import "github.com/os-artificer/AiT/internal/pkg/config"

// Public provider identifiers, re-exported from config so callers can write
// `provider.Deepseek` without importing config.
const (
	Deepseek = config.ProviderDeepseek
	Minimax  = config.ProviderMinimax
	Ollama   = config.ProviderOllama
)

// ID identifies an AI provider (matches config.Provider* values).
type ID = string
