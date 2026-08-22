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

// Package eino wraps CloudWeGo Eino ChatModel for AiT AI routing (Phase 0+ refactor).
//
// Phase 0 spike conclusions (2026-06-18):
//
//   - Strategy A confirmed: tiered Chat may use eino OpenAI-compatible models;
//     tiered Agent on Ollama stays on legacy endpoint until P4 specialized path.
//   - DeepSeek / MiniMax (openai api_style) / Ollama: eino-ext openai + base_url is sufficient for spike;
//     no dedicated deepseek component required yet.
//   - Anthropic Messages: no eino-ext anthropic component in tree; keep ShouldLegacyChat and
//     ShouldLegacyAgent (anthropic) on legacy HTTP until minimax_adapter or upstream component lands.
//   - FC tool_calls and OpenAI content XML (M4) post-process cleanly after eino Generate;
//     Anthropic XML (M6) uses the same agent.XMLToolCallToJSON on text blocks (see spike_test.go).
//
// This package must not import internal/ssh (product package aissh) or internal/pkg/ai/client.
package eino
