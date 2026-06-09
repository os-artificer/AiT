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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestTmuxOps_ChainModeUsesPTY (TM4) verifies chain mode uses PTY inject, not nil conn.
func TestTmuxOps_ChainModeUsesPTY(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	c.SetChainMode(true)

	err := c.tmuxEnsure(&bytes.Buffer{}, "ait-test-id")
	if err == nil {
		t.Fatal("expected inject error without live PTY markers")
	}
	if strings.Contains(err.Error(), "invalid memory address") || strings.Contains(err.Error(), "nil pointer") {
		t.Fatalf("tmuxEnsure used nil conn in chain mode: %v", err)
	}
}

// TestTmuxOps_DirectModeUsesConn verifies non-chain mode calls conn-based tmux (panics on nil conn).
func TestTmuxOps_DirectModeUsesConn(t *testing.T) {
	c := &Client{}
	c.SetChainMode(false)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic from nil conn in direct tmux path")
		}
	}()
	_ = c.tmuxEnsure(&bytes.Buffer{}, "ait-test-id")
}
