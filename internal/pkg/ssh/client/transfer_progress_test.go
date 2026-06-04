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
	if strings.Contains(got, "transferring") {
		t.Fatalf("expected copying label only: %q", got)
	}
}

func enableTransferProgressInlineTTY(t *testing.T) {
	t.Helper()
	orig := transferProgressInlineTTY
	transferProgressInlineTTY = func(*TransferProgress) bool { return true }
	t.Cleanup(func() { transferProgressInlineTTY = orig })
}

func TestTransferProgress_SFTPInlineTTY_Get(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.StartGet("/remote", "./local", -1)
	for i := int64(1); i <= 10; i++ {
		p.SFTPProgress(i*10, 100)
	}
	p.Done("done: get /remote → ./local (100 bytes)")

	got := buf.String()
	if strings.Count(got, "copying…") != 10 {
		t.Fatalf("expected 10 inline copying updates, got %q", got)
	}
	if strings.Contains(got, "transferring") {
		t.Fatalf("expected copying label: %q", got)
	}
	if strings.Count(got, termEOL) != 2 {
		t.Fatalf("expected start+done CRLF lines only, got %d in %q", strings.Count(got, termEOL), got)
	}
	if !strings.HasPrefix(got, "AiT:") {
		t.Fatalf("missing prefix: %q", got)
	}
}

func TestTransferProgress_SFTPInlineTTY_Put(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.StartPut("./local", "/remote", 100)
	for i := int64(1); i <= 5; i++ {
		p.SFTPProgress(i*20, 100)
	}
	p.Done("done: put ./local → /remote (100 bytes)")

	got := buf.String()
	if strings.Count(got, "copying…") != 5 {
		t.Fatalf("expected 5 inline copying updates, got %q", got)
	}
	if strings.Count(got, termEOL) != 2 {
		t.Fatalf("expected start+done CRLF lines only, got %d in %q", strings.Count(got, termEOL), got)
	}
}

func TestTransferProgress_InlineDoneCRLF(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.reportPercent(50)
	p.Done("done: put a → b (10 bytes)")

	got := buf.String()
	if !strings.Contains(got, "done: put a → b") {
		t.Fatalf("missing done line: %q", got)
	}
	if !strings.Contains(got, termEOL) {
		t.Fatalf("missing CRLF on done: %q", got)
	}
	if !strings.Contains(got, "\rAiT: copying") {
		t.Fatalf("expected inline progress with \\r: %q", got)
	}
}

// TestTransferProgress_SFTPClosingPhase (V6) reports closing… after copy completes.
func TestTransferProgress_SFTPClosingPhase(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.SFTPProgress(100, 100)
	p.PhaseClosing()

	got := buf.String()
	if !strings.Contains(got, "closing…") {
		t.Fatalf("missing closing phase: %q", got)
	}
	if strings.Contains(got, "\r\r\n") {
		t.Fatalf("double-wrapped CRLF: %q", got)
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
