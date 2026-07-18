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

package client

import "os"

const envRemoteSystemPrompt = "AIT_REMOTE_SYSTEM_PROMPT"

// UseRemoteSystemPrompt reports whether SSH interactive mode should keep the
// remote shell's native prompt unchanged instead of installing AiT prompt setup.
// Default is false (AiT-managed prompt). Set AIT_REMOTE_SYSTEM_PROMPT=1 to opt into
// native remote prompts.
func UseRemoteSystemPrompt() bool {
	return os.Getenv(envRemoteSystemPrompt) == "1"
}
