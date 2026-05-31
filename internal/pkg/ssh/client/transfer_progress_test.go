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

package client

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransferProgress_UsesCRLF(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.Phase("hashing…")

	got := buf.String()
	if !strings.Contains(got, termEOL) {
		t.Fatalf("expected CRLF line ending, got %q", got)
	}
	if strings.Contains(got, "AiT: AiT:") {
		t.Fatalf("double prefix: %q", got)
	}
}

func TestTransferProgress_WindowsPathDisplay(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	winPath := `C:\Users\me\file.txt`
	p.StartPut(winPath, "./remote.txt", 42)

	got := buf.String()
	if !strings.Contains(got, winPath) {
		t.Fatalf("expected Windows path in output, got %q", got)
	}
}

func TestTransferProgress_ChainModeOn(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.ChainModeOn()

	if !strings.Contains(buf.String(), "chain mode on") {
		t.Fatalf("missing chain line: %q", buf.String())
	}
}

func TestFormatTransferStart(t *testing.T) {
	got := formatTransferStart("get", "/remote", "./local", 86)
	if !strings.Contains(got, "86 bytes") {
		t.Fatalf("missing size: %q", got)
	}
}

func TestTransferProgress_DoneUsesPrefix(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.Done("done: ok")

	if !strings.HasPrefix(buf.String(), "AiT:") {
		t.Fatalf("missing AiT prefix: %q", buf.String())
	}
}

func TestTransferProgress_SFTPProgressPercent(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.SFTPProgress(50, 100)
	p.SFTPProgress(50, 100)
	p.SFTPProgress(100, 100)

	got := buf.String()
	if strings.Count(got, "50%") > 1 {
		t.Fatalf("duplicate 50%% reports: %q", got)
	}
	if !strings.Contains(got, "100%") {
		t.Fatalf("missing 100%%: %q", got)
	}
}

func TestTransferProgress_LocalPathFromFilepath(t *testing.T) {
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.StartGet("/remote", abs, 1)
	if !strings.Contains(buf.String(), abs) {
		t.Fatalf("expected abs path %q in %q", abs, buf.String())
	}
}
