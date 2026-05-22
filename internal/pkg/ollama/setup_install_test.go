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

package ollama

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
)

type fakeRunner struct {
	paths map[string]bool
	runs  []runRecord
}

type runRecord struct {
	name string
	args []string
}

func (f *fakeRunner) LookPath(name string) (string, error) {
	if f.paths[name] {
		return "/fake/" + name, nil
	}
	return "", errors.New("not found")
}

func (f *fakeRunner) Run(ctx context.Context, name string, args []string) error {
	f.runs = append(f.runs, runRecord{name: name, args: args})
	return nil
}

func TestEnsureBinary_SkippedWhenNotAllowed(t *testing.T) {
	r := &fakeRunner{paths: map[string]bool{}}
	if err := EnsureBinary(context.Background(), r, false); err != nil {
		t.Fatal(err)
	}
	if len(r.runs) != 0 {
		t.Fatalf("runs=%v", r.runs)
	}
}

func TestEnsureBinary_InstallWhenAllowed(t *testing.T) {
	r := &fakeRunner{paths: map[string]bool{}}
	switch runtime.GOOS {
	case "darwin":
		r.paths["brew"] = true
	case "linux":
		r.paths["curl"] = true
	case "windows":
		r.paths["winget"] = true
	default:
		t.Skip("unsupported GOOS")
	}

	if err := EnsureBinary(context.Background(), r, true); err != nil {
		t.Fatal(err)
	}
	if len(r.runs) != 1 {
		t.Fatalf("runs=%v", r.runs)
	}
	switch runtime.GOOS {
	case "darwin":
		if r.runs[0].name != "brew" || len(r.runs[0].args) < 2 || r.runs[0].args[0] != "install" {
			t.Fatalf("darwin run=%+v", r.runs[0])
		}
	case "linux":
		if r.runs[0].name != "sh" || !strings.Contains(r.runs[0].args[1], installScriptURL) {
			t.Fatalf("linux run=%+v", r.runs[0])
		}
	case "windows":
		if r.runs[0].name != "winget" || r.runs[0].args[0] != "install" {
			t.Fatalf("windows run=%+v", r.runs[0])
		}
	}
}

func TestEnsureBinary_AlreadyInstalled(t *testing.T) {
	r := &fakeRunner{paths: map[string]bool{binaryName: true, "brew": true}}
	if err := EnsureBinary(context.Background(), r, true); err != nil {
		t.Fatal(err)
	}
	if len(r.runs) != 0 {
		t.Fatalf("expected no install runs, got %v", r.runs)
	}
}
