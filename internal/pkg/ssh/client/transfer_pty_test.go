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
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestInjectCaptureRetry_SucceedsOnSecondAttempt(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer
	attempts := 0

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		attempts++
		if attempts == 1 {
			return "", ErrInjectCaptureNoMarkers
		}
		return "ylgeeker", nil
	}

	out, err := c.injectCaptureRetry(&stdin, "hostname")
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if out != "ylgeeker" {
		t.Fatalf("out = %q", out)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestInjectCaptureRetry_NonMarkerErrorStops(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer
	attempts := 0

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		attempts++
		return "", io.ErrClosedPipe
	}

	_, err := c.injectCaptureRetry(&stdin, "hostname")
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("expected closed pipe, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestWriteRemoteBase64_ChunkPhaseCount(t *testing.T) {
	encoded := strings.Repeat("A", ptyTransferChunkSize+10)
	chunks := chunkString(encoded, ptyTransferChunkSize)
	if len(chunks) != 2 {
		t.Fatalf("chunks = %d, want 2", len(chunks))
	}
}

func TestSettleShellAfterBurst_WritesTrue(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer

	if err := c.settleShellAfterBurst(&stdin); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdin.String(), transferShellSettleCmd) {
		t.Fatalf("stdin = %q, want true command", stdin.String())
	}
}

func TestHandleTransfer_SuccessPrintsDone(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.StartGet("/remote", "./local", 10)
	p.SetBytes(10)
	p.Done("done: get /remote → ./local (10 bytes)")

	got := buf.String()
	if !strings.Contains(got, "done:") {
		t.Fatalf("missing done line: %q", got)
	}
	if !strings.Contains(got, "10 bytes") {
		t.Fatalf("missing byte count: %q", got)
	}
}

func TestHandleTransfer_ChainOnShowsModeLine(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.ChainModeOn()
	if !strings.Contains(buf.String(), "transfer via PTY") {
		t.Fatalf("missing mode line: %q", buf.String())
	}
}

func TestPTYCaptureErrHint_IncludesPutUsage(t *testing.T) {
	err := wrapPTYCaptureErr(ErrInjectCaptureNoMarkers)
	s := err.Error()
	if !strings.Contains(s, "!ait put") {
		t.Fatalf("missing put hint: %q", s)
	}
	if !strings.Contains(s, "local-on-this-machine") {
		t.Fatalf("missing local path hint: %q", s)
	}
}

func TestPTYTransfer_WriteRemoteBase64_ReportsChunks(t *testing.T) {
	var phases []string
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	origPhase := transferProgressPhaseHook
	defer func() { transferProgressPhaseHook = origPhase }()
	transferProgressPhaseHook = func(p *TransferProgress, msg string) {
		phases = append(phases, msg)
	}

	c := &Client{scroll: &terminal.Buffer{}}
	origInject := ptyTransferInjectHook
	origSettle := ptyTransferSettleHook
	defer func() {
		ptyTransferInjectHook = origInject
		ptyTransferSettleHook = origSettle
	}()
	ptyTransferInjectHook = func(client *Client, stdin io.Writer, cmd string, opts InjectOpts) error {
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		return nil
	}

	pt := &PTYTransfer{client: c, progress: p}
	encoded := strings.Repeat("A", ptyTransferChunkSize+10)
	if err := pt.writeRemoteBase64(&bytes.Buffer{}, "/remote", encoded); err != nil {
		t.Fatal(err)
	}

	chunkLines := 0
	for _, line := range phases {
		if strings.Contains(line, "uploading chunk") {
			chunkLines++
		}
	}
	if chunkLines != 2 {
		t.Fatalf("chunk lines = %d, want 2, phases=%v", chunkLines, phases)
	}
}
