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
	"errors"
	"os"
	"strings"
	"testing"
)

func TestWrapLocalErr(t *testing.T) {
	err := wrapLocalErr("/tmp/missing", os.ErrNotExist)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "local file") || !strings.Contains(msg, "/tmp/missing") {
		t.Fatalf("message=%q", msg)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected ErrNotExist in chain: %v", err)
	}
}

func TestWrapRemoteErrNotExist(t *testing.T) {
	err := wrapRemoteErr("/temp/test", os.ErrNotExist)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "remote file") || !strings.Contains(msg, "/temp/test") {
		t.Fatalf("message=%q", msg)
	}
	if !strings.Contains(msg, "parent directory") {
		t.Fatalf("expected parent directory hint: %q", msg)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected ErrNotExist in chain: %v", err)
	}
}

func TestWrapRemoteErrOther(t *testing.T) {
	err := wrapRemoteErr("/a/b", os.ErrPermission)
	if !strings.Contains(err.Error(), "remote file") {
		t.Fatalf("message=%q", err.Error())
	}
	if strings.Contains(err.Error(), "parent directory") {
		t.Fatal("unexpected parent directory hint for permission error")
	}
}

func TestAbsLocalPath(t *testing.T) {
	dir := t.TempDir()
	rel := "f.txt"
	if err := os.WriteFile(dir+"/"+rel, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := absLocalPath(dir + "/" + rel)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, rel) {
		t.Fatalf("abs=%q", got)
	}
}
