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

package agentcore_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/approve"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/risk"
)

// TestPolicy_ReadAutoRun (AG-2) read commands auto-approve without prompting.
func TestPolicy_ReadAutoRun(t *testing.T) {
	t.Parallel()
	if risk.Classify("ls -la") != risk.Read {
		t.Fatal("ls should be Read")
	}
	st := approve.Trust{Policy: approve.PolicyPerCmd}
	ok, err := approve.Confirm(approve.Inputs{
		Trust:      &st,
		CmdLabel:   "run ls",
		PatternKey: "ls",
		Risk:       risk.Read,
		Policy:     approve.PolicyPerCmd,
	})
	if err != nil || !ok {
		t.Fatalf("Confirm=%v err=%v", ok, err)
	}
}

// TestPolicy_WriteConfirm (AG-3) low-risk write needs confirmation path.
func TestPolicy_WriteConfirm(t *testing.T) {
	t.Parallel()
	if risk.Classify("cp a b") != risk.Low {
		t.Fatalf("cp should be Low, got %v", risk.Classify("cp a b"))
	}
}

// TestPolicy_HighAlwaysConfirm (AG-4) high risk still confirms under session trust.
func TestPolicy_HighAlwaysConfirm(t *testing.T) {
	t.Parallel()
	st := approve.Trust{Policy: approve.PolicySessionOnce, SessionTrusted: true}
	ok, err := approve.Confirm(approve.Inputs{
		Trust:      &st,
		CmdLabel:   "rm -rf /tmp/x",
		PatternKey: "rm",
		Risk:       risk.High,
		Policy:     approve.PolicySessionOnce,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.ParseChoice("y", false, false), nil
		},
	})
	if err != nil || !ok {
		t.Fatalf("high risk y should confirm once: ok=%v err=%v", ok, err)
	}
}

// TestToolbox_ReadAutoRun (AG-5) toolbox read subcommands classify as Read.
func TestToolbox_ReadAutoRun(t *testing.T) {
	t.Parallel()
	if risk.ClassifyToolbox("aihost snapshot") != risk.Read {
		t.Fatal("aihost snapshot should be Read")
	}
}

// TestLoop_StepLimitEvent (AG-6) step limit event kind is stable for UI subscribers.
func TestLoop_StepLimitEvent(t *testing.T) {
	var buf bytes.Buffer
	emit := agentcore.FuncEmitter(func(e agentcore.Event) {
		if e.Kind == agentcore.EventStepLimitReached {
			buf.WriteString(e.Detail)
		}
	})
	emit(agentcore.Event{Kind: agentcore.EventStepLimitReached, Detail: "12"})
	if !strings.Contains(buf.String(), "12") {
		t.Fatalf("got %q", buf.String())
	}
}
