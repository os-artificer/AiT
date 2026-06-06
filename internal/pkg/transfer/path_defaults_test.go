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

package transfer

import (
	"path/filepath"
	"testing"
)

// TestDefaultLocalPathForGet (TR5) verifies standalone/meta default local basename.
func TestDefaultLocalPathForGet(t *testing.T) {
	if got := DefaultLocalPathForGet("/remote/path/file"); got != "./file" {
		t.Fatalf("got %q want ./file", got)
	}
	if got := DefaultLocalPathForGet("./bin/ait-darwin-arm64"); got != "./ait-darwin-arm64" {
		t.Fatalf("got %q want ./ait-darwin-arm64", got)
	}
}

// TestLocalPathWithDirDefault (TR4/TR5) appends basename when local is a directory.
func TestLocalPathWithDirDefault(t *testing.T) {
	isDir := func(p string) bool { return p == "./out/" || p == "./out" }
	got := LocalPathWithDirDefault("/remote/file", "./out/", isDir)
	want := filepath.Join("./out", "file")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if LocalPathWithDirDefault("/remote/file", "./dest", isDir) != "./dest" {
		t.Fatal("expected unchanged file path")
	}
}

// TestDefaultRemotePathForPut (TR9) verifies default remote basename for put.
func TestDefaultRemotePathForPut(t *testing.T) {
	if got := DefaultRemotePathForPut("./bin/app"); got != "./app" {
		t.Fatalf("got %q want ./app", got)
	}
	if got := DefaultRemotePathForPut("/local/path/file"); got != "./file" {
		t.Fatalf("got %q want ./file", got)
	}
}

// TestRemotePathWithDirDefault (TR8/TR9) appends basename when remote is a directory.
func TestRemotePathWithDirDefault(t *testing.T) {
	isDir := func(p string) bool { return p == "./remote-dir/" || p == "./remote-dir" }
	got := RemotePathWithDirDefault("./local/file", "./remote-dir/", isDir)
	want := "./remote-dir/file"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if RemotePathWithDirDefault("./local/file", "./dest", isDir) != "./dest" {
		t.Fatal("expected unchanged file path")
	}
}
