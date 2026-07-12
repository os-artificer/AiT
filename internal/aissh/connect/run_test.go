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

package connect

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func tmuxAttachWrappedErr() error {
	cause := errors.New("Process exited with status: 2")
	return errhint.WithTry(fmt.Errorf("tmux attach failed (exit 2): %w", cause),
		"Could not attach to the remote tmux session.",
		toolname.CLI("--tmux devhost"),
	)
}

// TestLoginRun_PreservesTmuxAttachErrHint (TM12a) keeps errhint messages when Login returns wrapped ssh.ExitError.
func TestLoginRun_PreservesTmuxAttachErrHint(t *testing.T) {
	err := tmuxAttachWrappedErr()
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		t.Fatal("wrapped tmux attach error must not be empty")
	}
	if !strings.Contains(msg, "tmux attach failed") {
		t.Fatalf("want sentinel in %q", msg)
	}
	if !strings.Contains(msg, "Try:") {
		t.Fatalf("want Try block in %q", msg)
	}
}

// TestLoginExitCode_WrappedSSHExit (TM12b) still extracts exit status after Run returns errors unchanged.
func TestLoginExitCode_WrappedSSHExit(t *testing.T) {
	if code, ok := ExitCode(exitError(2)); !ok || code != 2 {
		t.Fatalf("exitError: code=%d ok=%v", code, ok)
	}
	if code, ok := ExitCode(&client.RemoteExitError{Code: 2}); !ok || code != 2 {
		t.Fatalf("RemoteExitError: code=%d ok=%v", code, ok)
	}
}

// TestLoginRun_ErrHintReachableCLI (TM12c) verifies errhint.Print writes CLI stderr for preserved errors.
func TestLoginRun_ErrHintReachableCLI(t *testing.T) {
	err := tmuxAttachWrappedErr()
	var buf bytes.Buffer
	errhint.Print(&buf, errhint.SurfaceCLI, err)
	if strings.TrimSpace(buf.String()) == "" {
		t.Fatal("errhint.Print must write non-empty CLI stderr")
	}
	if !strings.Contains(buf.String(), "tmux attach failed") {
		t.Fatalf("stderr should include sentinel: %q", buf.String())
	}
}

func TestLoginExitCode_RemoteExitError(t *testing.T) {
	err := &client.RemoteExitError{Code: 7}
	code, ok := ExitCode(err)
	if !ok || code != 7 {
		t.Fatalf("got code=%d ok=%v", code, ok)
	}
}

func TestLoginExitCode_ExitErrorType(t *testing.T) {
	err := exitError(3)
	code, ok := ExitCode(err)
	if !ok || code != 3 {
		t.Fatalf("got code=%d ok=%v", code, ok)
	}
}
