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
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
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

func TestPTYLeakErrHint_IncludesInspectAdvice(t *testing.T) {
	err := wrapPTYLeakErr(ErrInjectLeakSignature)
	s := err.Error()
	if !strings.Contains(s, "leaked PTY inject") {
		t.Fatalf("missing leak explanation: %q", s)
	}
	if !strings.Contains(s, "cat/head") {
		t.Fatalf("missing inspect hint: %q", s)
	}
}

func TestInject_SuppressEchoChainQuiet(t *testing.T) {
	c := &Client{}
	if got := injectQuietForClient(c, InjectOpts{SuppressEcho: true}); got != scrollback.CaptureQuiet {
		t.Fatalf("default suppress quiet = %v, want %v", got, scrollback.CaptureQuiet)
	}
	c.SetChainMode(true)
	if got := injectQuietForClient(c, InjectOpts{SuppressEcho: true}); got != transferChainCaptureQuiet {
		t.Fatalf("chain suppress quiet = %v, want %v", got, transferChainCaptureQuiet)
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
	origScript := ptyTransferScriptHook
	origSettle := ptyTransferSettleHook
	origMktemp := ptyTransferMktempHook
	origVerify := ptyTransferVerifyHook
	origCapture := injectCaptureForRetry
	defer func() {
		ptyTransferScriptHook = origScript
		ptyTransferSettleHook = origSettle
		ptyTransferMktempHook = origMktemp
		ptyTransferVerifyHook = origVerify
		injectCaptureForRetry = origCapture
	}()
	ptyTransferMktempHook = func(client *Client, stdin io.Writer) (string, error) {
		return "/tmp/ait-put.test", nil
	}
	ptyTransferScriptHook = func(client *Client, stdin io.Writer, script string) error {
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		return nil
	}
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.HasPrefix(cmd, "head -c ") {
			return base64.StdEncoding.EncodeToString([]byte("payload")), nil
		}
		return "", ErrInjectCaptureNoMarkers
	}
	ptyTransferVerifyHook = func(client *Client, stdin io.Writer, remotePath string) (string, error) {
		return "abc123  /remote\n", nil
	}

	pt := &PTYTransfer{client: c, progress: p}
	encoded := strings.Repeat("A", ptyTransferChunkSize+10)
	hash, err := pt.writeRemoteBase64(&bytes.Buffer{}, "/remote", encoded, int64(len(encoded)))
	if err != nil {
		t.Fatal(err)
	}
	if hash != "abc123" {
		t.Fatalf("hash = %q, want abc123", hash)
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

func TestPTYPut_WritesViaTempFile(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var scripts []string
	var mvSeen bool

	origScript := ptyTransferScriptHook
	origSettle := ptyTransferSettleHook
	origMktemp := ptyTransferMktempHook
	origVerify := ptyTransferVerifyHook
	origCapture := injectCaptureForRetry
	defer func() {
		ptyTransferScriptHook = origScript
		ptyTransferSettleHook = origSettle
		ptyTransferMktempHook = origMktemp
		ptyTransferVerifyHook = origVerify
		injectCaptureForRetry = origCapture
	}()

	tmpPath := "/tmp/ait-put.mock"
	ptyTransferMktempHook = func(client *Client, stdin io.Writer) (string, error) {
		return tmpPath, nil
	}
	ptyTransferScriptHook = func(client *Client, stdin io.Writer, script string) error {
		scripts = append(scripts, script)
		if strings.HasPrefix(strings.TrimSpace(script), "mv ") {
			mvSeen = true
		}
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		return nil
	}
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.HasPrefix(cmd, "head -c ") {
			return base64.StdEncoding.EncodeToString([]byte("hello")), nil
		}
		return "", ErrInjectCaptureNoMarkers
	}
	ptyTransferVerifyHook = func(client *Client, stdin io.Writer, remotePath string) (string, error) {
		return localFileHash([]byte("hello")) + "  ./remote\n", nil
	}

	dir := t.TempDir()
	localPath := dir + "/local"
	if err := os.WriteFile(localPath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	pt := &PTYTransfer{client: c, progress: NewTransferProgress(io.Discard)}
	if err := pt.Put(t.Context(), &bytes.Buffer{}, localPath, "./remote"); err != nil {
		t.Fatal(err)
	}
	if !mvSeen {
		t.Fatal("expected mv script to remote path")
	}
	for _, script := range scripts {
		if strings.Contains(script, ">> './remote'") || strings.Contains(script, "> './remote'") {
			t.Fatalf("chunk must not redirect to final remote path: %q", script)
		}
		if strings.Contains(script, ptyPutChunkHeredocDelim) && !strings.Contains(script, shellQuoteArg(tmpPath)) {
			t.Fatalf("heredoc chunk must append to tmp only: %q", script)
		}
	}
}

func TestWriteRemoteBase64_SettlesBetweenSteps(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	settleCount := 0

	origScript := ptyTransferScriptHook
	origSettle := ptyTransferSettleHook
	origMktemp := ptyTransferMktempHook
	origVerify := ptyTransferVerifyHook
	origCapture := injectCaptureForRetry
	defer func() {
		ptyTransferScriptHook = origScript
		ptyTransferSettleHook = origSettle
		ptyTransferMktempHook = origMktemp
		ptyTransferVerifyHook = origVerify
		injectCaptureForRetry = origCapture
	}()

	ptyTransferMktempHook = func(client *Client, stdin io.Writer) (string, error) {
		return "/tmp/ait-put.test", nil
	}
	ptyTransferScriptHook = func(client *Client, stdin io.Writer, script string) error {
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		settleCount++
		return nil
	}
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.HasPrefix(cmd, "head -c ") {
			return base64.StdEncoding.EncodeToString([]byte("x")), nil
		}
		return "", ErrInjectCaptureNoMarkers
	}
	ptyTransferVerifyHook = func(client *Client, stdin io.Writer, remotePath string) (string, error) {
		return "deadbeef  /remote\n", nil
	}

	pt := &PTYTransfer{client: c}
	encoded := strings.Repeat("A", ptyTransferChunkSize+10)
	if _, err := pt.writeRemoteBase64(&bytes.Buffer{}, "/remote", encoded, int64(1)); err != nil {
		t.Fatal(err)
	}

	// 2 chunks + mv => 3 settle calls between script steps.
	if settleCount != 3 {
		t.Fatalf("settleCount = %d, want 3", settleCount)
	}
}

func TestPTYPut_AbortsOnInjectLeakSignature(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	verifyCalled := false

	origScript := ptyTransferScriptHook
	origSettle := ptyTransferSettleHook
	origMktemp := ptyTransferMktempHook
	origVerify := ptyTransferVerifyHook
	origCapture := injectCaptureForRetry
	defer func() {
		ptyTransferScriptHook = origScript
		ptyTransferSettleHook = origSettle
		ptyTransferMktempHook = origMktemp
		ptyTransferVerifyHook = origVerify
		injectCaptureForRetry = origCapture
	}()

	ptyTransferMktempHook = func(client *Client, stdin io.Writer) (string, error) {
		return "/tmp/ait-put.test", nil
	}
	ptyTransferScriptHook = func(client *Client, stdin io.Writer, script string) error {
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		return nil
	}
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.HasPrefix(cmd, "head -c ") {
			leaked := injectEchoOff + " printf junk"
			return base64.StdEncoding.EncodeToString([]byte(leaked)), nil
		}
		return "", ErrInjectCaptureNoMarkers
	}
	ptyTransferVerifyHook = func(client *Client, stdin io.Writer, remotePath string) (string, error) {
		verifyCalled = true
		return "abc  /remote\n", nil
	}

	dir := t.TempDir()
	localPath := dir + "/local"
	if err := os.WriteFile(localPath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	pt := &PTYTransfer{client: c, progress: NewTransferProgress(io.Discard)}
	err := pt.Put(t.Context(), &bytes.Buffer{}, localPath, "./remote")
	if !errors.Is(err, ErrInjectLeakSignature) {
		t.Fatalf("expected ErrInjectLeakSignature, got %v", err)
	}
	if verifyCalled {
		t.Fatal("verify must not run when leak guard fails")
	}
}

func TestPTYPut_VerifyViaSettleCapture(t *testing.T) {
	var phases []string
	origPhase := transferProgressPhaseHook
	defer func() { transferProgressPhaseHook = origPhase }()
	transferProgressPhaseHook = func(p *TransferProgress, msg string) {
		phases = append(phases, msg)
	}

	c := &Client{scroll: &terminal.Buffer{}}
	origScript := ptyTransferScriptHook
	origSettle := ptyTransferSettleHook
	origMktemp := ptyTransferMktempHook
	origVerify := ptyTransferVerifyHook
	origCapture := injectCaptureForRetry
	defer func() {
		ptyTransferScriptHook = origScript
		ptyTransferSettleHook = origSettle
		ptyTransferMktempHook = origMktemp
		ptyTransferVerifyHook = origVerify
		injectCaptureForRetry = origCapture
	}()
	ptyTransferMktempHook = func(client *Client, stdin io.Writer) (string, error) {
		return "/tmp/ait-put.test", nil
	}
	ptyTransferScriptHook = func(client *Client, stdin io.Writer, script string) error {
		return nil
	}
	ptyTransferSettleHook = func(client *Client, stdin io.Writer) error {
		return nil
	}
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.HasPrefix(cmd, "head -c ") {
			return base64.StdEncoding.EncodeToString([]byte("hello")), nil
		}
		return "", ErrInjectCaptureNoMarkers
	}
	verifyCalled := false
	ptyTransferVerifyHook = func(client *Client, stdin io.Writer, remotePath string) (string, error) {
		verifyCalled = true
		return localFileHash([]byte("hello")) + "  ./remote\n", nil
	}

	dir := t.TempDir()
	localPath := dir + "/local"
	if err := os.WriteFile(localPath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	pt := &PTYTransfer{client: c, progress: NewTransferProgress(io.Discard)}
	if err := pt.Put(t.Context(), &bytes.Buffer{}, localPath, "./remote"); err != nil {
		t.Fatal(err)
	}
	if !verifyCalled {
		t.Fatal("expected settle+capture verify hook")
	}
	foundVerify := false
	for _, p := range phases {
		if strings.Contains(p, "verifying") {
			foundVerify = true
		}
	}
	if !foundVerify {
		t.Fatalf("missing verifying phase: %v", phases)
	}
}

func TestCaptureQuiet_ChainMode(t *testing.T) {
	c := &Client{}
	if got := captureQuietForClient(c); got != scrollback.CaptureQuiet {
		t.Fatalf("default quiet = %v, want %v", got, scrollback.CaptureQuiet)
	}
	c.SetChainMode(true)
	if got := captureQuietForClient(c); got != transferChainCaptureQuiet {
		t.Fatalf("chain quiet = %v, want %v", got, transferChainCaptureQuiet)
	}
}

func TestInjectCaptureRetry_ChainAttempts(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	c.SetChainMode(true)
	var stdin bytes.Buffer
	attempts := 0

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		attempts++
		return "", ErrInjectCaptureNoMarkers
	}

	_, err := c.injectCaptureRetry(&stdin, "sha256sum '/tmp/x'")
	if !errors.Is(err, ErrInjectCaptureNoMarkers) {
		t.Fatalf("expected marker error, got %v", err)
	}
	if attempts != injectCaptureRetryAttemptsChain {
		t.Fatalf("attempts = %d, want %d", attempts, injectCaptureRetryAttemptsChain)
	}
}
