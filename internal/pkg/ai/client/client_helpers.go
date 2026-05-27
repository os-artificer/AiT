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

import (
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
)

// ProviderDisabledErr returns a user-visible error for a disabled provider override.
func ProviderDisabledErr(id string) error {
	return provider.DisabledErr(id)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	i := len(s) - max
	for i < len(s) && !utf8.RuneStart(s[i]) {
		i++
	}
	return s[i:]
}
