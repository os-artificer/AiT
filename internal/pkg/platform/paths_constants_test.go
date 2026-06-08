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

package platform

import (
	"strings"
	"testing"
)

// TestConfigFileName_FromPlatform (C17) keeps the config basename contract stable.
func TestConfigFileName_FromPlatform(t *testing.T) {
	if ConfigFileName != "config.yaml" {
		t.Fatalf("ConfigFileName=%q", ConfigFileName)
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, ConfigFileName) {
		t.Fatalf("ConfigPath suffix: %q", path)
	}
}
