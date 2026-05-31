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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
	pkgtransfer "github.com/os-artificer/AiT/internal/pkg/transfer"
	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

func TestZModemSession_ExclusiveIO(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer

	_, cleanup, err := c.BeginZModemSession(true, &stdin, t.TempDir()+"/out", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if !c.zmodemActive.Load() {
		t.Fatal("expected zmodem active")
	}

	c.feedZModemOutput([]byte("test payload"))
	if strings.Contains(c.scroll.String(), "test payload") {
		t.Fatal("zmodem bytes should not enter scrollback")
	}
}

func TestZModemSession_StripProtocolNoise(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte("noise**\x18B0more"))

	var stdin bytes.Buffer
	_, cleanup, err := c.BeginZModemSession(true, &stdin, t.TempDir()+"/out", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()

	if strings.Contains(c.scroll.String(), "**\x18B0") {
		t.Fatal("expected zmodem noise stripped from scrollback")
	}
}

func TestZModemTransfer_MissingLrzsz(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		return "missing", nil
	}

	zt := &ZModemTransfer{client: c, progress: NewTransferProgress(io.Discard)}
	err := zt.Get(t.Context(), &bytes.Buffer{}, "/remote", "./local")
	if !errors.Is(err, ErrMissingLrzsz) {
		t.Fatalf("expected ErrMissingLrzsz, got %v", err)
	}
}

func TestHandleTransferPTY_UsesZModem(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	c.SetChainMode(true)

	origProbe := injectCaptureForRetry
	defer func() { injectCaptureForRetry = origProbe }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			return "ok", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	dir := t.TempDir()
	local := dir + "/local"
	if err := os.WriteFile(local, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	origWait := zmodemTransferWaitHook
	defer func() { zmodemTransferWaitHook = origWait }()
	zmodemTransferWaitHook = func(ctx context.Context, sess *zmodem.Session) error {
		return fmt.Errorf("zmodem: no protocol response from remote peer")
	}

	err := c.handleTransferPTY(t.Context(), &bytes.Buffer{}, pkgtransfer.DirPut, []string{local, "./remote"}, NewTransferProgress(io.Discard))
	if err == nil {
		t.Fatal("expected zmodem put to fail without live PTY peer")
	}
	if !strings.Contains(err.Error(), "zmodem") {
		t.Fatalf("expected zmodem error, got %v", err)
	}
}

func TestZModemTransfer_ProgressCRLF(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.Phase("handshake…")
	p.Phase("transferring… 50%")
	if !strings.Contains(buf.String(), "\r\n") {
		t.Fatalf("missing CRLF progress: %q", buf.String())
	}
}

func TestPTYZModemErrHint_IncludesLrzsz(t *testing.T) {
	err := wrapPTYZModemErr(ErrMissingLrzsz)
	s := err.Error()
	if !strings.Contains(s, "lrzsz") {
		t.Fatalf("missing lrzsz hint: %q", s)
	}
}
