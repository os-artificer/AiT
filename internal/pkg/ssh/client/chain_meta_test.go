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
	"io"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func withFastPwdProbe(t *testing.T) {
	t.Helper()
	orig := probeShellPwdHook
	probeShellPwdHook = func(_ *Client, _ io.Writer) (string, error) {
		return "", nil
	}
	t.Cleanup(func() { probeShellPwdHook = orig })
}

// TestHandleMetaLine_ChainLogin (N1) verifies chain login suspends and returns ChainActionError.
func TestHandleMetaLine_ChainLogin(t *testing.T) {
	withFastPwdProbe(t)

	var suspendCalled bool
	c := &Client{scroll: &terminal.Buffer{}}
	err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, "!ait user@hostB", InteractiveOpts{
		OnSuspend: func(_ *terminal.Buffer, _ string, _, _ int) error {
			suspendCalled = true
			return nil
		},
	})
	if err == nil {
		t.Fatal("expected chain action error")
	}
	var chain *ChainActionError
	if !errors.As(err, &chain) {
		t.Fatalf("expected ChainActionError, got %T: %v", err, err)
	}
	if chain.Op != ChainOpLogin {
		t.Fatalf("op = %v", chain.Op)
	}
	if !suspendCalled {
		t.Fatal("OnSuspend not called")
	}
}

// TestHandleMetaLine_ChainLoginOnSuspendFail keeps session when suspend fails.
func TestHandleMetaLine_ChainLoginOnSuspendFail(t *testing.T) {
	withFastPwdProbe(t)

	c := &Client{scroll: &terminal.Buffer{}}
	want := io.EOF
	err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, "!ait user@hostB", InteractiveOpts{
		OnSuspend: func(_ *terminal.Buffer, _ string, _, _ int) error {
			return want
		},
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want suspend error", err)
	}
}
