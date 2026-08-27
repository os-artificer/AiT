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
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryPath_UnderDataDir(t *testing.T) {
	isolatePlatformHome(t)

	data, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	path, err := HistoryPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, data) {
		t.Fatalf("path %q not under DataDir %q", path, data)
	}
	if filepath.Base(path) != "repl_history" {
		t.Fatalf("basename = %q, want repl_history", filepath.Base(path))
	}
}
