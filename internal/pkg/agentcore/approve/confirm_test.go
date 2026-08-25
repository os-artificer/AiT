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

package approve

import (
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/risk"
)

// A21b: RenderPrompt produces newlines that, after CRLF conversion, contain the
// Choice menu starting at column 0.
func TestRenderPrompt_LowRiskMenu(t *testing.T) {
	p := RenderPrompt("local", "run hostname", risk.Low)
	if !strings.Contains(p, "\n") {
		t.Fatal("expected newlines in prompt template")
	}
	if !strings.Contains(p, "Choice [y/c/s/N]:") {
		t.Fatalf("missing choice line: %q", p)
	}
}

// A21d: High-risk prompts must omit the c/s options.
func TestRenderPrompt_HighRiskOmitsPatternSession(t *testing.T) {
	p := RenderPrompt("local", "rm -rf /tmp/x", risk.High)
	if strings.Contains(p, "[c]") || strings.Contains(p, "trust this command pattern") {
		t.Fatalf("high risk should not offer pattern/session: %q", p)
	}
	if !strings.Contains(p, "Choice [y/N]:") {
		t.Fatal("expected y/N only")
	}
	if !strings.Contains(p, "HIGH-RISK") {
		t.Fatal("expected HIGH-RISK label")
	}
}

// A21c: ParseChoice maps y/c/s correctly and respects allowPattern/allowSession gating.
func TestParseChoice_Variants(t *testing.T) {
	res := ParseChoice("s", true, true)
	if !res.Run || !res.TrustSession || res.TrustPattern {
		t.Fatalf("got %+v", res)
	}
	res = ParseChoice("y", true, true)
	if !res.Run || res.TrustSession || res.TrustPattern {
		t.Fatalf("y should run once only: %+v", res)
	}
	res = ParseChoice("c", true, true)
	if !res.Run || !res.TrustPattern || res.TrustSession {
		t.Fatalf("c should trust pattern: %+v", res)
	}
	res = ParseChoice("s", false, false)
	if res.Run {
		t.Fatal("s should not run when allowSession=false (high risk)")
	}
	res = ParseChoice("n", true, true)
	if res.Run {
		t.Fatal("n should not run")
	}
	res = ParseChoice("", true, true)
	if res.Run {
		t.Fatal("empty should not run")
	}
}

// A21f: PatternKey returns the first token; PatternKeyAit prefixes ait:.
func TestPatternKey_FirstToken(t *testing.T) {
	if PatternKey("hostname -I") != "hostname" {
		t.Fatal("hostname -I")
	}
	if PatternKey("hostnamectl") != "hostnamectl" {
		t.Fatal("hostnamectl")
	}
	if PatternKey("") != "" {
		t.Fatal("empty")
	}
	if PatternKeyAit("list") != "ait:list" {
		t.Fatal("ait:list")
	}
	if PatternKeyAit("  ") != "" {
		t.Fatal("ait empty")
	}
}

// Trust.IsPatternTrusted handles empty key and nil receiver gracefully.
func TestTrust_IsPatternTrusted(t *testing.T) {
	var nilT *Trust
	if nilT.IsPatternTrusted("hostname") {
		t.Fatal("nil receiver must return false")
	}
	tr := &Trust{}
	if tr.IsPatternTrusted("") {
		t.Fatal("empty key must return false")
	}
	tr.AddTrustedPattern("hostname")
	if !tr.IsPatternTrusted("hostname") {
		t.Fatal("expected trusted")
	}
}

// Trust.ApplyResult promotes Policy when TrustSession is set.
func TestTrust_ApplyResult_PromotesPolicy(t *testing.T) {
	tr := &Trust{Policy: PolicyPerCmd}
	tr.ApplyResult(Result{Run: true, TrustSession: true}, "")
	if !tr.SessionTrusted {
		t.Fatal("expected SessionTrusted")
	}
	if tr.Policy != PolicySessionOnce {
		t.Fatal("expected Policy promoted to PolicySessionOnce")
	}
}

// Trust.Clear resets session/pattern trust but keeps Policy.
func TestTrust_Clear_KeepsPolicy(t *testing.T) {
	tr := &Trust{
		Policy:          PolicySessionOnce,
		SessionTrusted:  true,
		TrustedPatterns: map[string]struct{}{"hostname": {}},
	}
	tr.Clear()
	if tr.SessionTrusted || tr.TrustedPatterns != nil {
		t.Fatal("clear must drop session/pattern trust")
	}
	if tr.Policy != PolicySessionOnce {
		t.Fatal("clear must keep Policy")
	}
}

// A20f: Risk.High ignores session trust and re-prompts (empty answer → no run).
func TestConfirm_HighRiskIgnoresSession(t *testing.T) {
	tr := &Trust{Policy: PolicySessionOnce, SessionTrusted: true}
	ok, err := Confirm(Inputs{
		Trust:      tr,
		ScopeLabel: "local",
		CmdLabel:   "rm -rf /tmp/x",
		PatternKey: "rm",
		Risk:       risk.High,
		Policy:     PolicySessionOnce,
		Out:        io.Discard,
		In:         strings.NewReader(""),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("empty confirm should not approve high risk")
	}
}

// Risk.Read auto-approves without prompting, regardless of policy.
func TestConfirm_ReadAutoApproves(t *testing.T) {
	n := 0
	tr := &Trust{Policy: PolicyPerCmd}
	ok, err := Confirm(Inputs{
		Trust:       tr,
		ScopeLabel:  "local",
		CmdLabel:    "ls -la",
		PatternKey:  "ls",
		Risk:        risk.Read,
		Policy:      PolicyPerCmd,
		ConfirmRead: func(string) (Result, error) { n++; return Result{}, nil },
		Out:         io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("read commands should auto-approve")
	}
	if n != 0 {
		t.Fatalf("confirm was called %d times, want 0", n)
	}
}

// A20g: Once the user picks "s" for a low-risk command, the next low-risk command
// auto-approves without asking again.
func TestConfirm_ApproveSessionLowRiskSkipsSecond(t *testing.T) {
	n := 0
	tr := &Trust{Policy: PolicySessionOnce}
	in := Inputs{
		Trust:      tr,
		ScopeLabel: "local",
		CmdLabel:   "hostname",
		PatternKey: "hostname",
		Risk:       risk.Low,
		Policy:     PolicySessionOnce,
		Out:        io.Discard,
		ConfirmRead: func(string) (Result, error) {
			n++
			return Result{Run: true, TrustSession: true}, nil
		},
	}
	if ok, err := Confirm(in); err != nil || !ok {
		t.Fatalf("first ok=%v err=%v", ok, err)
	}
	if !tr.SessionTrusted {
		t.Fatal("should trust after session choice")
	}
	if ok, err := Confirm(in); err != nil || !ok {
		t.Fatalf("second ok=%v err=%v", ok, err)
	}
	if n != 1 {
		t.Fatalf("confirm asked %d times, want 1", n)
	}
}

// A21f: TrustPattern records the first token; subsequent commands with the same
// first token skip prompts. Different first tokens still require confirmation.
func TestConfirm_PatternTrustFirstToken(t *testing.T) {
	tr := &Trust{}
	first := Inputs{
		Trust:      tr,
		ScopeLabel: "local",
		CmdLabel:   "hostname -I",
		PatternKey: "hostname",
		Risk:       risk.Low,
		ConfirmRead: func(string) (Result, error) {
			return Result{Run: true, TrustPattern: true}, nil
		},
	}
	if ok, err := Confirm(first); err != nil || !ok {
		t.Fatalf("first ok=%v err=%v", ok, err)
	}

	n := 0
	second := Inputs{
		Trust:      tr,
		ScopeLabel: "local",
		CmdLabel:   "hostname",
		PatternKey: "hostname",
		Risk:       risk.Low,
		ConfirmRead: func(string) (Result, error) {
			n++
			return Result{}, nil
		},
	}
	if ok, err := Confirm(second); err != nil || !ok {
		t.Fatalf("hostname trusted should auto-approve: ok=%v err=%v", ok, err)
	}
	if n != 0 {
		t.Fatalf("confirm called %d times for trusted pattern", n)
	}

	third := second
	third.CmdLabel = "hostnamectl"
	third.PatternKey = "hostnamectl"
	if ok, _ := Confirm(third); ok {
		t.Fatal("hostnamectl should still require confirm")
	}
}
