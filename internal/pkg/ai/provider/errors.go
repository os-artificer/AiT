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

package provider

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// ErrNotConfigured is the sentinel when no AI client is available.
var ErrNotConfigured = errhint.NewSentinel("AI not configured")

// NotConfiguredErr returns a multi-line hint for missing AI configuration.
func NotConfiguredErr() error {
	return errhint.WithTry(ErrNotConfigured,
		"Set a cloud provider API key or enable tiered Ollama routing.",
		"ait config set ai.deepseek.api_key <key>",
		"ait config set ai.minimax.api_key <key>",
		"ait ai provider enable ollama",
		"ait ai --help",
	)
}

// ErrChatMessageRequired is the sentinel when chat has no message.
var ErrChatMessageRequired = errhint.NewSentinel("chat requires a message")

// ChatMessageRequiredErr returns a multi-line hint for empty chat input.
func ChatMessageRequiredErr() error {
	return errhint.WithTry(ErrChatMessageRequired,
		"Provide a message on the command line or via stdin.",
		"ait ai chat <message>",
		"echo ping | ait ai chat",
		"!ait ai chat <message>",
		"ait ai chat --help",
	)
}

// DisabledErr reports a disabled provider override.
func DisabledErr(id string) error {
	return errhint.WithTryMsg(fmt.Sprintf("provider %q is disabled", id),
		"Enable the provider before using it as --model override.",
		fmt.Sprintf("ait ai provider enable %s", id),
		"ait ai --help",
	)
}
