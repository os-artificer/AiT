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
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

func TestDial_DefaultChainModeOff(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	if c.ChainMode() {
		t.Fatal("expected chain mode off by default")
	}
}

func TestHandleTransfer_UsesSFTPWhenChainOff(t *testing.T) {
	c := &Client{opts: Options{Host: "bastion"}, scroll: &terminal.Buffer{}}
	usePTY, auto, err := c.resolveTransferMode(nil)
	if err != nil {
		t.Fatalf("resolveTransferMode: %v", err)
	}
	if usePTY || auto {
		t.Fatalf("expected SFTP path, got usePTY=%v auto=%v", usePTY, auto)
	}
}

func TestHandleTransfer_UsesPTYWhenChainOn(t *testing.T) {
	c := &Client{opts: Options{Host: "bastion"}, scroll: &terminal.Buffer{}}
	c.SetChainMode(true)
	usePTY, auto, err := c.resolveTransferMode(nil)
	if err != nil {
		t.Fatalf("resolveTransferMode: %v", err)
	}
	if !usePTY || auto {
		t.Fatalf("expected PTY path, got usePTY=%v auto=%v", usePTY, auto)
	}
}

func TestInjectCapture_MissingMarkersError(t *testing.T) {
	delta := "prompt noise without capture markers"
	if captureMarkersPresent(delta) {
		t.Fatal("expected no markers")
	}

	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte(delta))
	if captureMarkersPresent(c.scroll.String()) {
		t.Fatal("scroll should not contain markers")
	}
	err := fmt.Errorf("%w", ErrInjectCaptureNoMarkers)
	if !errors.Is(err, ErrInjectCaptureNoMarkers) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestInjectCapture_StripsScrollback(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	leaked := "noise" + injectCaptureBegin + "data" + injectCaptureEnd + injectCaptureBeginPrint
	_, _ = c.scroll.Write([]byte(leaked))

	c.scroll.StripSubstrings(injectStripFragments()...)
	remaining := c.scroll.String()
	if strings.Contains(remaining, injectCaptureBegin) {
		t.Fatalf("begin marker still in scrollback: %q", remaining)
	}
	if strings.Contains(remaining, injectCaptureEnd) {
		t.Fatalf("end marker still in scrollback: %q", remaining)
	}
	if strings.Contains(remaining, injectCaptureBeginPrint) {
		t.Fatalf("begin print still in scrollback: %q", remaining)
	}
}

func TestHandleTransferPTY_PausesRemoteOut(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := c.handleTransferPTY(ctx, &bytes.Buffer{}, transfer.DirGet, []string{"/remote", "/local"}); err == nil {
		t.Fatal("expected context error")
	}
	if c.pauseRemoteOut.Load() {
		t.Fatal("pauseRemoteOut should be restored after handleTransferPTY")
	}
}

func TestNormalizeHost_ShortNameAndFQDN(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"bastion", "bastion.example.com", true},
		{"Bastion.Example.COM", "bastion", true},
		{"user@bastion", "bastion.local", true},
		{"bastion:22", "bastion", true},
		{"bastion", "internal", false},
	}
	for _, tc := range tests {
		got := hostsMatch(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("hostsMatch(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestResolveTransferMode_AutoEnablesChain(t *testing.T) {
	if !activeShellNeedsChain("bastion", "internal-server") {
		t.Fatal("expected chain needed for different hosts")
	}
	if activeShellNeedsChain("bastion.example.com", "bastion") {
		t.Fatal("same host short/FQDN should not need chain")
	}
	if activeShellNeedsChain("bastion", "") {
		t.Fatal("empty active host should not need chain")
	}
}

func TestTransferWrongHopHint_Format(t *testing.T) {
	cause := fmt.Errorf("remote file %q: %w", "/path", os.ErrNotExist)
	err := wrapTransferWrongHop("bastion", cause)
	if !errors.Is(err, ErrTransferWrongHop) {
		t.Fatalf("expected ErrTransferWrongHop, got %v", err)
	}
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, "!ait chain on") {
		t.Fatalf("missing chain on hint: %q", s)
	}
	if !strings.Contains(s, "directly connected host") {
		t.Fatalf("missing explanation: %q", s)
	}
}

func TestPTYCaptureErrHint_Format(t *testing.T) {
	err := wrapPTYCaptureErr(ErrInjectCaptureNoMarkers)
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if !strings.Contains(s, "which base64 sha256sum") {
		t.Fatalf("missing tool hint: %q", s)
	}
	if !strings.Contains(s, "!ait chain off") {
		t.Fatalf("missing chain off hint: %q", s)
	}
}

func TestCaptureMarkersPresent(t *testing.T) {
	okDelta := "x" + injectCaptureBegin + "payload" + injectCaptureEnd
	if !captureMarkersPresent(okDelta) {
		t.Fatal("expected markers present")
	}
	if captureMarkersPresent("no markers here") {
		t.Fatal("expected markers absent")
	}
	if captureMarkersPresent(injectCaptureBegin + "partial") {
		t.Fatal("expected false when end marker missing")
	}
}

func TestIsSFTPNotFoundErr(t *testing.T) {
	if !isSFTPNotFoundErr(os.ErrNotExist) {
		t.Fatal("expected not exist")
	}
	wrapped := fmt.Errorf("remote file %q: %w (file or parent directory may not exist; check the remote path)", "/x", os.ErrNotExist)
	if !isSFTPNotFoundErr(wrapped) {
		t.Fatal("expected wrapped not exist")
	}
	if isSFTPNotFoundErr(fmt.Errorf("permission denied")) {
		t.Fatal("unexpected match")
	}
}
