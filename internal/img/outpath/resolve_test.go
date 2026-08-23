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

package outpath_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/img/outpath"
)

func TestAutoName_Collision(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(src, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(dir, "photo_aiimg.jpg")
	if err := os.WriteFile(first, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := outpath.AutoName(src)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "photo_aiimg_1.jpg")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestOutpath_ConfirmN(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(src, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var prompt bytes.Buffer
	tty := true
	dsts, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{src},
		PromptOut: &prompt,
		IsTTY:     &tty,
		Confirm: func(p string) (bool, error) {
			_, _ = prompt.WriteString(p)
			return false, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dsts) != 1 || !strings.HasSuffix(dsts[0], "photo_aiimg.jpg") {
		t.Fatalf("dsts=%v", dsts)
	}
	if !strings.Contains(prompt.String(), "Overwrite original file(s)?") {
		t.Fatalf("prompt=%q", prompt.String())
	}
}

func TestOutpath_ConfirmY(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(src, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var prompt bytes.Buffer
	tty := true
	dsts, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{src},
		PromptOut: &prompt,
		IsTTY:     &tty,
		Confirm:   func(string) (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dsts) != 1 || dsts[0] != src {
		t.Fatalf("dsts=%v", dsts)
	}
}

func TestResolve_OutputEmpty(t *testing.T) {
	_, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{"a.jpg"},
		OutputSet: true,
		Output:    "",
	})
	if !errors.Is(err, outpath.ErrOutputEmpty) {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_OutputIsDir(t *testing.T) {
	dir := t.TempDir()
	_, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{"a.jpg"},
		OutputSet: true,
		Output:    dir,
	})
	if !errors.Is(err, outpath.ErrOutputIsDir) {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_OutputMultiFile(t *testing.T) {
	_, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{"a.jpg", "b.jpg"},
		OutputSet: true,
		Output:    "out.jpg",
	})
	if !errors.Is(err, outpath.ErrOutputMultiFile) {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_OutputInplaceConflict(t *testing.T) {
	_, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     []string{"a.jpg"},
		OutputSet: true,
		Output:    "out.jpg",
		Inplace:   true,
	})
	if !errors.Is(err, outpath.ErrOutputInplaceConflict) {
		t.Fatalf("got %v", err)
	}
}
