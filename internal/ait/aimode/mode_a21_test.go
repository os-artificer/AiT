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

package aimode

import (
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

func TestPrompt_ShowsInteractionMode(t *testing.T) {
	if !strings.Contains(Prompt(ai.ScopeLocal, "", ModeAgent), "(Agent)") {
		t.Fatal("expected Agent in prompt")
	}
	if !strings.Contains(Prompt(ai.ScopeLocal, "", ModeAsk), "(Ask)") {
		t.Fatal("expected Ask in prompt")
	}
}

func TestConfirmPrompt_UsesCRLF(t *testing.T) {
	p := confirmPromptRisk(Options{Scope: ai.ScopeLocal}, "run hostname", RiskLow)
	if !strings.Contains(p, "\n") {
		t.Fatal("expected newlines in prompt template")
	}
	var buf strings.Builder
	aimWrite(&buf, p)
	if !strings.Contains(buf.String(), "\r\n") {
		t.Fatalf("expected CRLF output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Choice [y/c/s/N]:") {
		t.Fatalf("missing choice line: %q", buf.String())
	}
}

func TestParseConfirm_SessionTrust(t *testing.T) {
	res := parseConfirmChoice("s", true, true)
	if !res.Run || !res.TrustSession || res.TrustPattern {
		t.Fatalf("got %+v", res)
	}
	res = parseConfirmChoice("y", true, true)
	if !res.Run || res.TrustSession || res.TrustPattern {
		t.Fatalf("y should run once only: %+v", res)
	}
	res = parseConfirmChoice("c", true, true)
	if !res.Run || !res.TrustPattern || res.TrustSession {
		t.Fatalf("c should trust pattern: %+v", res)
	}
}

func TestConfirm_SessionTrustFromPrompt(t *testing.T) {
	st := newSessionState(Options{})
	opt := Options{
		Scope: ai.ScopeLocal,
		ConfirmRead: func(string) (ConfirmResult, error) {
			return ConfirmResult{Run: true, TrustSession: true}, nil
		},
	}
	ok, err := confirmExecRisk(opt, st, "hostname", "hostname", RiskLow)
	if err != nil || !ok || !st.sessionTrusted {
		t.Fatalf("ok=%v err=%v trusted=%v", ok, err, st.sessionTrusted)
	}
}

func TestCommandPatternKey_FirstToken(t *testing.T) {
	if CommandPatternKey("hostname -I") != "hostname" {
		t.Fatal("hostname -I")
	}
	if CommandPatternKey("hostnamectl") != "hostnamectl" {
		t.Fatal("hostnamectl")
	}
	if CommandPatternKeyAit("list") != "ait:list" {
		t.Fatal("ait:list")
	}
}

func TestConfirm_PatternTrustSkipsSamePrefix(t *testing.T) {
	st := newSessionState(Options{})
	opt := Options{
		Scope: ai.ScopeLocal,
		ConfirmRead: func(string) (ConfirmResult, error) {
			return ConfirmResult{Run: true, TrustPattern: true}, nil
		},
	}
	ok, err := confirmExecRisk(opt, st, "hostname -I", "hostname", RiskLow)
	if err != nil || !ok {
		t.Fatal(err)
	}
	n := 0
	opt.ConfirmRead = func(string) (ConfirmResult, error) {
		n++
		return ConfirmResult{}, nil
	}
	ok, err = confirmExecRisk(opt, st, "hostname", "hostname", RiskLow)
	if err != nil || !ok {
		t.Fatalf("second hostname should skip confirm: ok=%v err=%v", ok, err)
	}
	if n != 0 {
		t.Fatalf("confirm called %d times", n)
	}
	ok, err = confirmExecRisk(opt, st, "hostnamectl", "hostnamectl", RiskLow)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("hostnamectl should still require confirm")
	}
}

func TestConfirm_HighRiskNoSessionOption(t *testing.T) {
	p := confirmPromptRisk(Options{Scope: ai.ScopeLocal}, "rm -rf /tmp/x", RiskHigh)
	if strings.Contains(p, "[c]") || strings.Contains(p, "trust this command pattern") {
		t.Fatalf("high risk should not offer pattern/session: %q", p)
	}
	if !strings.Contains(p, "Choice [y/N]:") {
		t.Fatal("expected y/N only")
	}
	res := parseConfirmChoice("s", false, false)
	if res.Run {
		t.Fatal("s should not run on high risk menu")
	}
}

func TestSetApprove_EachClearsPatterns(t *testing.T) {
	st := &sessionState{
		sessionTrusted:  true,
		trustedPatterns: map[string]struct{}{"hostname": {}},
		policy:          ExecPolicySessionOnce,
	}
	if err := setApprove(st, "each", io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.sessionTrusted || st.trustedPatterns != nil {
		t.Fatal("each should clear trust")
	}
}
