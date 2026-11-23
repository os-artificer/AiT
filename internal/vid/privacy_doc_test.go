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

package aivid_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// VD16: guide mentions privacy / local-only / best-effort plates.
func TestGuide_AitvidPrivacy(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	body, err := os.ReadFile(filepath.Join(root, "docs/guide/aitvid.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"隐私", "--local-only", "尽力而为"} {
		if !strings.Contains(s, want) {
			t.Fatalf("guide missing %q", want)
		}
	}
}
