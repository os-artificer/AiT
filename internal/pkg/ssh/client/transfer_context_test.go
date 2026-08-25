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
	"context"
	"io"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/pty"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestProcessForwardStdin_TransferCtrlC (V7) cancels transfer without writing ^C to remote stdin.
func TestProcessForwardStdin_TransferCtrlC(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	pr, pw := io.Pipe()
	c.stdinReader = pty.StartFrom(pr)

	var stdin bytes.Buffer
	transferCtx, end := c.beginTransfer(context.Background())
	defer end()
	defer func() { _ = pw.Close() }()

	done := make(chan struct{})
	go func() {
		<-transferCtx.Done()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	if _, err := pw.Write([]byte{transferInterruptByte}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("transfer context not cancelled")
	}

	if stdin.Len() != 0 {
		t.Fatalf("stdin should not receive ^C during transfer, got %q", stdin.Bytes())
	}
}

// TestProcessForwardStdin_CtrlCFwdWhenInactive forwards ^C when no transfer is active.
func TestProcessForwardStdin_CtrlCFwdWhenInactive(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer

	res := c.processForwardStdinByte(
		context.Background(),
		&stdin,
		stdinByte{B: transferInterruptByte},
		newMetaScanner(""),
		InteractiveOpts{},
	)
	if res.action == fwdStdinReturn && res.err != nil {
		t.Fatalf("unexpected return: %v", res.err)
	}
	if stdin.Len() != 1 || stdin.Bytes()[0] != transferInterruptByte {
		t.Fatalf("expected ^C forwarded, stdin=%q", stdin.Bytes())
	}
}
