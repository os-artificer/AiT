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

package config

import (
	"testing"

	"github.com/os-artificer/AiT/internal/ssh/config/shared"
)

func TestParseAPIKeyConfigKey_P4(t *testing.T) {
	cases := []struct {
		key      string
		provider string
		ok       bool
	}{
		{"deepseek.api_key", "deepseek", true},
		{"minimax.api_key", "minimax", true},
		{"ai.deepseek.api_key", "deepseek", true},
		{"ai.minimax.api_key", "minimax", true},
		{"ai.default", "", false},
	}
	for _, tc := range cases {
		p, ok := shared.ParseAPIKeyConfigKey(tc.key)
		if ok != tc.ok || p != tc.provider {
			t.Errorf("ParseAPIKeyConfigKey(%q) = %q, %v want %q, %v", tc.key, p, ok, tc.provider, tc.ok)
		}
	}
}
