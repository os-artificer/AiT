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
	"strings"
	"testing"
)

func TestUnknownConfigKeyErr_Format(t *testing.T) {
	err := unknownConfigKeyErr("bad.key")
	s := err.Error()
	if !strings.Contains(s, "unknown config key: bad.key") {
		t.Fatalf("missing summary: %q", s)
	}
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, "ait config set --help") {
		t.Fatalf("missing config help try line: %q", s)
	}
}
