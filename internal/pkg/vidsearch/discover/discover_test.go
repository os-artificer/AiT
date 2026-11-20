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

package discover_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/discover"
)

func TestListVideos_DirAndRecursive(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.mp4"), "x")
	mustWrite(t, filepath.Join(root, "b.txt"), "x")
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(sub, "c.mov"), "x")

	got, err := discover.ListVideos(root, discover.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("non-recursive want 1 got %v", got)
	}

	got, err = discover.ListVideos(root, discover.Options{Recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("recursive want 2 got %v", got)
	}
}

func TestListVideos_Empty(t *testing.T) {
	root := t.TempDir()
	_, err := discover.ListVideos(root, discover.Options{})
	if !errors.Is(err, discover.ErrNoVideos) {
		t.Fatalf("got %v", err)
	}
}

func TestListVideos_MaxVideos(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "1.mp4"), "x")
	mustWrite(t, filepath.Join(root, "2.mp4"), "x")
	mustWrite(t, filepath.Join(root, "3.mp4"), "x")
	_, err := discover.ListVideos(root, discover.Options{MaxVideos: 2})
	if !errors.Is(err, discover.ErrTooManyVideos) {
		t.Fatalf("got %v", err)
	}
}

func TestListVideos_UnlimitedDefault(t *testing.T) {
	root := t.TempDir()
	const n = 55
	for i := 0; i < n; i++ {
		mustWrite(t, filepath.Join(root, fmt.Sprintf("v%d.mp4", i)), "x")
	}
	got, err := discover.ListVideos(root, discover.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != n {
		t.Fatalf("want %d got %d", n, len(got))
	}
	got, err = discover.ListVideos(root, discover.Options{MaxVideos: 0})
	if err != nil || len(got) != n {
		t.Fatalf("MaxVideos=0: len=%d err=%v", len(got), err)
	}
	got, err = discover.ListVideos(root, discover.Options{MaxVideos: -1})
	if err != nil || len(got) != n {
		t.Fatalf("MaxVideos=-1: len=%d err=%v", len(got), err)
	}
}

func TestListVideos_ExplicitOddExt(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "clip.bin")
	mustWrite(t, p, "x")
	got, err := discover.ListVideos(p, discover.Options{})
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v err %v", got, err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
