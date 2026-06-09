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
	"errors"
	"strings"
	"testing"
)

// TestWrapTmuxAttachExitErr_Format (TM10c) verifies attach failures include sentinel and Try block.
func TestWrapTmuxAttachExitErr_Format(t *testing.T) {
	cause := errors.New("Process exited with status: 2")
	err := wrapTmuxAttachExitErr("ait-id-1", 2, cause)
	msg := err.Error()

	if !strings.Contains(msg, errTmuxAttachFailed) {
		t.Fatalf("want sentinel %q in %q", errTmuxAttachFailed, msg)
	}
	if !strings.Contains(msg, "Try:") {
		t.Fatalf("want Try block in %q", msg)
	}
	if !strings.Contains(msg, "ait login --tmux devhost") {
		t.Fatalf("want tmux login try line in %q", msg)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error should preserve cause for errors.Is")
	}
}

func TestWrapTmuxPreAttachErr_Format(t *testing.T) {
	cause := errors.New("run-shell failed")
	err := wrapTmuxPreAttachErr("ait-id-1", cause)
	msg := err.Error()

	if !strings.Contains(msg, errTmuxPaneSetupFailed) {
		t.Fatalf("want sentinel %q in %q", errTmuxPaneSetupFailed, msg)
	}
	if !strings.Contains(msg, "Try:") {
		t.Fatalf("want Try block in %q", msg)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error should preserve cause")
	}
}

func TestWrapTmuxSessionNotFoundErr_Format(t *testing.T) {
	err := wrapTmuxSessionNotFoundErr("ait-id-1", nil)
	msg := err.Error()

	if !strings.Contains(msg, errTmuxSessionNotFound) {
		t.Fatalf("want sentinel %q in %q", errTmuxSessionNotFound, msg)
	}
	if !strings.Contains(msg, "Try:") {
		t.Fatalf("want Try block in %q", msg)
	}
}

func TestWrapInteractiveWaitErr_PassthroughNonTmux(t *testing.T) {
	cause := errors.New("shell exited")
	if got := wrapInteractiveWaitErr(false, "ait-id-1", cause); got != cause {
		t.Fatalf("non-tmux wait should return cause unchanged, got %v", got)
	}
}

// TestWrapTmuxAttachExitErr_PreservesCause (TM12b) keeps the wait error chain for login ExitCode/errors.Is.
func TestWrapTmuxAttachExitErr_PreservesCause(t *testing.T) {
	cause := errors.New("Process exited with status: 2")
	got := wrapTmuxAttachExitErr("ait-id-1", 2, cause)
	if !errors.Is(got, cause) {
		t.Fatal("wrapped error should preserve cause for errors.Is")
	}
	if !strings.Contains(got.Error(), errTmuxAttachFailed) {
		t.Fatalf("want sentinel in %q", got.Error())
	}
}
