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

package aissh

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestConfigInitCmd(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "init"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Created") {
		t.Fatalf("output=%q", out.String())
	}

	out.Reset()
	err = ExecuteWriters([]string{"config", "init"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Fatalf("second init output=%q", out.String())
	}
}

func TestConfigInitCmd_DryRun(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "init", "--dry-run"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Would create") {
		t.Fatalf("output=%q", out.String())
	}
	if _, err := os.Stat(root + "/config.yaml"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("dry-run should not write config.yaml")
	}
}
