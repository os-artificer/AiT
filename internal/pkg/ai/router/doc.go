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

// Package router contains the receiver-free building blocks of the AiT AI
// multi-provider router: the failover Chain of Responsibility helpers
// (ResolveTryOrder, AggregateErrors, LogFailover, APIKeyHint, IsRetryable)
// and the tiered routing classifier (ClassifyPrompt) shared by the cloud and
// local-first paths.
//
// The Strategy choice (OpenAI vs Anthropic API style per provider endpoint)
// stays in the client subpackage because the strategy implementations are
// receiver methods on the (unexported) endpoint type.
package router
