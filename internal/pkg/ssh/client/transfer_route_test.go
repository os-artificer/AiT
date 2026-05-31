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

	"github.com/os-artificer/AiT/internal/pkg/lrzsz"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/transfer/route"
)

func TestRoute_DirectUsesSFTPWhenLrzszPresent(t *testing.T) {
	c := &Client{
		opts:   Options{Host: "bastion", LoginDestination: "bastion"},
		scroll: &terminal.Buffer{},
	}
	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "hostname") {
			return "bastion", nil
		}
		if strings.Contains(cmd, "command -v sz") {
			return "ok", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	plan, _, _, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Backend != route.BackendSFTP || !plan.LrzszOK {
		t.Fatalf("expected direct SFTP when lrzsz present, got %+v", plan)
	}
	if plan.ChainOnly {
		t.Fatal("direct host should not be chain-only")
	}
}

func TestRoute_ChainPrefersZModemWhenLrzszPresent(t *testing.T) {
	c := &Client{
		opts:   Options{Host: "bastion"},
		scroll: &terminal.Buffer{},
	}
	c.SetChainMode(true)

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			return "ok", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	plan, _, _, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Backend != route.BackendZModem || !plan.LrzszOK {
		t.Fatalf("expected chain zmodem plan, got %+v", plan)
	}
	if !plan.ChainOnly {
		t.Fatal("expected chain-only")
	}
}

func TestRoute_FallsBackSFTPDirect(t *testing.T) {
	c := &Client{
		opts:   Options{Host: "bastion"},
		scroll: &terminal.Buffer{},
	}
	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "hostname") {
			return "bastion", nil
		}
		if strings.Contains(cmd, "command -v sz") {
			return "missing", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	plan, _, _, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Backend != route.BackendSFTP || plan.LrzszOK {
		t.Fatalf("expected SFTP fallback, got %+v", plan)
	}
}

func TestRoute_ChainDeclinedInstall(t *testing.T) {
	c := &Client{
		opts:   Options{Host: "bastion"},
		scroll: &terminal.Buffer{},
	}
	c.SetChainMode(true)

	origConfirm := transferInstallConfirmHook
	defer func() { transferInstallConfirmHook = origConfirm }()
	transferInstallConfirmHook = func(prompt string) (bool, error) {
		return false, nil
	}

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			return "missing", nil
		}
		if strings.Contains(cmd, "uname") || strings.Contains(cmd, "apt-get") {
			return "linux-apt", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	plan, host, target, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.ChainOnly || plan.LrzszOK {
		t.Fatalf("unexpected plan: %+v", plan)
	}

	err = c.ensureLrzszForChain(t.Context(), &bytes.Buffer{}, host, target)
	if !errors.Is(err, lrzsz.ErrInstallDeclined) {
		t.Fatalf("expected declined, got %v", err)
	}
}

func TestRoute_ChainInstallThenZModem(t *testing.T) {
	c := &Client{
		opts:   Options{Host: "bastion"},
		scroll: &terminal.Buffer{},
	}
	c.SetChainMode(true)

	origConfirm := transferInstallConfirmHook
	defer func() { transferInstallConfirmHook = origConfirm }()
	transferInstallConfirmHook = func(prompt string) (bool, error) {
		return true, nil
	}

	probeCount := 0
	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			probeCount++
			if probeCount >= 2 {
				return "ok", nil
			}
			return "missing", nil
		}
		if strings.Contains(cmd, "uname") || strings.Contains(cmd, "apt-get") {
			return "linux-apt", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	plan, host, target, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.ChainOnly {
		t.Fatal("expected chain")
	}

	err = c.ensureLrzszForChain(t.Context(), &bytes.Buffer{}, host, target)
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	ok, err := c.probeLrzsz(&bytes.Buffer{})
	if err != nil || !ok {
		t.Fatalf("probe after install: ok=%v err=%v", ok, err)
	}
}

func TestLrzszResolveInstall_NonTTY(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	c.SetChainMode(true)

	origConfirm := transferInstallConfirmHook
	defer func() { transferInstallConfirmHook = origConfirm }()
	transferInstallConfirmHook = nil

	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			return "missing", nil
		}
		return "linux-apt", nil
	}

	_, _, target, err := c.planTransfer(&bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}

	origOpts := transferInstallOptsFromClient
	defer func() { transferInstallOptsFromClient = origOpts }()
	transferInstallOptsFromClient = func(_ *Client) transferInstallOpts {
		return transferInstallOpts{In: strings.NewReader("")}
	}

	err = c.ensureLrzszForChain(t.Context(), &bytes.Buffer{}, "internal", target)
	if !errors.Is(err, lrzsz.ErrNonInteractiveInstall) {
		t.Fatalf("expected non-interactive install err, got %v", err)
	}
}
