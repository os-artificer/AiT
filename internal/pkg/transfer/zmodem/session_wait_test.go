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

package zmodem

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	xzmodem "github.com/os-artificer/AiT/pkg/third-party/xiwh/zmodem/zmodem"
)

// TestSession_Wait_TeardownAfterDone (V1) verifies idle is reached soon after data done.
func TestSession_Wait_TeardownAfterDone(t *testing.T) {
	sess, err := NewReceiveSession(io.Discard, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.started.Store(true)
	sess.finish(nil)

	if err := sess.Wait(t.Context()); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

// TestSession_Wait_TeardownTimeout (V2) returns ErrTeardownTimeout when idle never follows data done.
func TestSession_Wait_TeardownTimeout(t *testing.T) {
	orig := teardownTimeoutDuration
	teardownTimeoutDuration = 50 * time.Millisecond
	defer func() { teardownTimeoutDuration = orig }()

	sess, err := NewReceiveSession(io.Discard, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.started.Store(true)
	sess.testBlockIdle.Store(true)
	sess.finish(nil)

	err = sess.Wait(t.Context())
	if err == nil {
		t.Fatal("expected teardown timeout")
	}
	if !errors.Is(err, ErrTeardownTimeout) {
		t.Fatalf("expected ErrTeardownTimeout, got %v", err)
	}
}

// TestSession_Wait_ContextCancelAborts (V8) aborts when ctx is cancelled after data starts.
func TestSession_Wait_ContextCancelAborts(t *testing.T) {
	var ptyOut captureWriter
	sess, err := NewReceiveSession(&ptyOut, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.started.Store(true)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	start := time.Now()
	err = sess.Wait(ctx)
	if time.Since(start) > 2*time.Second {
		t.Fatal("Wait blocked too long after cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if len(ptyOut) == 0 {
		t.Fatal("expected abort sequence on cancel")
	}
}

// TestSession_Wait_TeardownAfterZFINWithPrefix (V10) completes teardown after PTY noise + ZEOF/ZFIN.
func TestSession_Wait_TeardownAfterZFINWithPrefix(t *testing.T) {
	zeof, err := xzmodem.HexFrameBytes(xzmodem.ZEOF, xzmodem.DEFAULT_HEADER_DATA)
	if err != nil {
		t.Fatal(err)
	}
	zfin, err := xzmodem.HexFrameBytes(xzmodem.ZFIN, xzmodem.DEFAULT_HEADER_DATA)
	if err != nil {
		t.Fatal(err)
	}

	var ptyOut captureWriter
	sess, err := NewReceiveSession(&ptyOut, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.engine.EnterReceiveForTest()
	sess.started.Store(true)

	errCh := make(chan error, 1)
	go func() {
		errCh <- sess.Wait(t.Context())
	}()

	time.Sleep(30 * time.Millisecond)
	sess.finish(nil)

	prefix := []byte("Bytes/sec 9999\r\n")
	payload := append(append(append([]byte(nil), prefix...), zeof...), zfin...)
	if _, err := sess.Feed(payload); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Wait: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait timed out after ZFIN with prefix")
	}
}

// TestSession_Wait_SendTeardownAfterUpload (V13) signals dataDone at 100% and completes after ZFIN.
func TestSession_Wait_SendTeardownAfterUpload(t *testing.T) {
	zfin, err := xzmodem.HexFrameBytes(xzmodem.ZFIN, xzmodem.DEFAULT_HEADER_DATA)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	path := dir + "/payload"
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	var ptyOut captureWriter
	sess, err := NewSendSession(&ptyOut, path, "remote.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.engine.EnterSendForTest()
	sess.started.Store(true)

	errCh := make(chan error, 1)
	go func() {
		errCh <- sess.Wait(t.Context())
	}()

	time.Sleep(30 * time.Millisecond)
	sess.reportSendUploadProgress(nil, 4, 4)

	if _, err := sess.Feed(zfin); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Wait: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait timed out after send upload and ZFIN")
	}
}

// TestSession_Wait_TeardownChainTimeout (V12) uses the chain teardown deadline when enabled.
func TestSession_Wait_TeardownChainTimeout(t *testing.T) {
	origChain := teardownTimeoutChainDuration
	teardownTimeoutChainDuration = 80 * time.Millisecond
	defer func() { teardownTimeoutChainDuration = origChain }()

	sess, err := NewReceiveSession(io.Discard, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.SetChainTeardown(true)
	sess.engine.EnterReceiveForTest()
	sess.started.Store(true)
	sess.testBlockIdle.Store(true)
	sess.finish(nil)

	start := time.Now()
	err = sess.Wait(t.Context())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected teardown timeout")
	}
	if !errors.Is(err, ErrTeardownTimeout) {
		t.Fatalf("expected ErrTeardownTimeout, got %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("chain teardown took %v; expected ~%v", elapsed, teardownTimeoutChainDuration)
	}
}

type captureWriter []byte

func (c *captureWriter) Write(p []byte) (int, error) {
	*c = append(*c, p...)
	return len(p), nil
}
