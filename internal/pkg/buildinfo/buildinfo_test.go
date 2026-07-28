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

package buildinfo_test

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/buildinfo"
)

func TestFormat_LdflagsWin(t *testing.T) {
	got := buildinfo.Format(buildinfo.Inputs{
		Version:   "v1.2.3",
		GitCommit: "abcdef123456",
		BuildDate: "2024-01-02T03:04:05Z",
		Embedded:  "v0.0.0",
	})
	if !strings.Contains(got, "Version:   v1.2.3") {
		t.Fatalf("%q", got)
	}
	if !strings.Contains(got, "GitHash:   abcdef123456") {
		t.Fatalf("%q", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Fatal("trailing newline")
	}
}

func TestShort_DevFallback(t *testing.T) {
	if got := buildinfo.Short(buildinfo.Inputs{}); got != "dev" {
		t.Fatalf("%q", got)
	}
}

func TestShort_Embedded(t *testing.T) {
	if got := buildinfo.Short(buildinfo.Inputs{Embedded: " v9.9.9 \n"}); got != "v9.9.9" {
		t.Fatalf("%q", got)
	}
}
