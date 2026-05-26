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

package ai

import "github.com/os-artificer/AiT/internal/pkg/ai/provider"

// ErrNotConfigured is the sentinel when no AI client is available (alias).
var ErrNotConfigured = provider.ErrNotConfigured

// NotConfiguredErr returns a multi-line hint for missing AI configuration.
func NotConfiguredErr() error {
	return provider.NotConfiguredErr()
}

// ErrChatMessageRequired is the sentinel when chat has no message (alias).
var ErrChatMessageRequired = provider.ErrChatMessageRequired

// ChatMessageRequiredErr returns a multi-line hint for empty chat input.
func ChatMessageRequiredErr() error {
	return provider.ChatMessageRequiredErr()
}

// ProviderDisabledErr reports a disabled provider override.
func ProviderDisabledErr(id string) error {
	return provider.DisabledErr(id)
}
