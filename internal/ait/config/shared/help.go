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

package shared

// KeysHelp is the shared configuration keys section for config subcommand help.
const KeysHelp = `  session.auth_ttl              duration for session unlock (e.g. 168h, 7d)
  session.remote_prompt.path    short (last path segment) or absolute (full path, ~ for home)
  ai.default                    default AI provider (deepseek|minimax)
  ai.fallback                   comma-separated failover order (optional)
  ai.providers.<id>.base_url    provider API base URL
  ai.providers.<id>.model       model name
  ai.providers.<id>.api_style   openai or anthropic (minimax only for anthropic)
  ai.<provider>.api_key         API key in OS keyring (recommended)
  deepseek.api_key              alias for ai.deepseek.api_key
  minimax.api_key               alias for ai.minimax.api_key
  deepseek.base_url / .model    aliases for ai.providers.deepseek.*
`
