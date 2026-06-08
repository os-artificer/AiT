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

package dirlist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirlist_FormatMode(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := WriteLongListing(&buf, root); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "drwx") {
		t.Fatalf("want directory mode prefix drwx, got %q", out)
	}
	if !strings.Contains(out, "-rw-r--r--") {
		t.Fatalf("want file mode prefix -rw-r--r--, got %q", out)
	}
}

func TestDirlist_SymlinkSuffix(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("t"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "alink")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Skip("symlink not supported:", err)
	}

	var buf bytes.Buffer
	if err := WriteLongListing(&buf, root); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "alink"+symlinkArrow) {
		t.Fatalf("want symlink arrow in output: %q", buf.String())
	}
}

func TestWriteLongListing_HasDotEntries(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	if err := WriteLongListing(&buf, root); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "total ") {
		t.Fatalf("want total line: %q", out)
	}
	for _, want := range []string{" .\n", " ..\n"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}
