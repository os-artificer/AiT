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
	"testing"
	"time"
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

type captureWriter []byte

func (c *captureWriter) Write(p []byte) (int, error) {
	*c = append(*c, p...)
	return len(p), nil
}
