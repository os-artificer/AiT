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
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/risk"
)

// Result is the parsed user answer to a confirmation prompt.
type Result struct {
	// Run indicates whether the command should execute.
	Run bool
	// TrustPattern indicates the user pressed "c" to trust the first-token pattern.
	TrustPattern bool
	// TrustSession indicates the user pressed "s" to trust safe commands for the session.
	TrustSession bool
}

// Inputs collects everything Confirm needs without depending on product CLI packages.
type Inputs struct {
	// Trust mutates as the user makes choices; must be non-nil for stateful confirms.
	Trust *Trust
	// ScopeLabel is the bracketed scope string in the prompt header ("local" or "remote").
	ScopeLabel string
	// CmdLabel is the command text shown to the user.
	CmdLabel string
	// PatternKey is the first-token key for TrustPattern propagation.
	PatternKey string
	// Risk classifies the command; Interactive short-circuits to (false, nil).
	Risk risk.Level
	// Policy is the current session policy.
	Policy Policy
	// ConfirmRead is the preferred path; when non-nil it replaces Out/In prompt+read.
	ConfirmRead func(prompt string) (Result, error)
	// Out receives the prompt text when ConfirmRead is nil.
	Out io.Writer
	// In supplies one line of user input when ConfirmRead is nil.
	In io.Reader
	// Write writes prompt to Out; pass a CRLF writer to convert "\n" to "\r\n" for SSH PTY.
	// When nil, io.WriteString is used directly.
	Write func(io.Writer, string)
}

// Confirm runs the confirmation flow and returns true when the user approved the command.
// It mutates in.Trust on TrustPattern / TrustSession choices.
// Risk.Interactive returns (false, nil) without prompting.
// Risk.Read returns (true, nil) without prompting — read-only commands are auto-approved.
// Returns an error only when the underlying I/O fails.
func Confirm(in Inputs) (bool, error) {
	if in.Risk == risk.Interactive {
		return false, nil
	}
	if in.Risk == risk.Read {
		return true, nil
	}

	if in.Risk != risk.High && in.Trust.IsPatternTrusted(in.PatternKey) {
		return true, nil
	}
	if in.Risk != risk.High && in.Policy == PolicySessionOnce && in.Trust != nil && in.Trust.SessionTrusted {
		return true, nil
	}

	allowPattern := in.Risk != risk.High
	allowSession := in.Risk != risk.High
	prompt := RenderPrompt(in.ScopeLabel, in.CmdLabel, in.Risk)

	if in.ConfirmRead != nil {
		res, err := in.ConfirmRead(prompt)
		if err != nil {
			return false, err
		}
		if !res.Run {
			return false, nil
		}
		in.Trust.ApplyResult(res, in.PatternKey)
		return true, nil
	}

	if in.Write != nil {
		in.Write(in.Out, prompt)
	} else {
		_, _ = io.WriteString(in.Out, prompt)
	}
	line, err := readLine(in.In)
	if err != nil {
		return false, err
	}

	res := ParseChoice(line, allowPattern, allowSession)
	if !res.Run {
		return false, nil
	}
	in.Trust.ApplyResult(res, in.PatternKey)
	return true, nil
}

// ParseChoice interprets one confirmation line (y/c/s/n) with allowPattern/allowSession
// gating "c"/"s" for high-risk prompts that omit those options.
// Empty / unknown input returns the zero Result (do not run).
func ParseChoice(line string, allowPattern, allowSession bool) Result {
	line = strings.ToLower(strings.TrimSpace(line))
	switch line {
	case "y", "yes":
		return Result{Run: true}
	case "c", "cmd", "pattern":
		if !allowPattern {
			return Result{}
		}
		return Result{Run: true, TrustPattern: true}
	case "s", confirmKeySession, confirmKeyTrust:
		if !allowSession {
			return Result{}
		}
		return Result{Run: true, TrustSession: true}
	default:
		return Result{}
	}
}

// RenderPrompt formats the multi-line confirmation prompt for a given scope label,
// command text, and risk level. High-risk prompts omit the c/s options.
// Output uses "\n" separators; callers convert to CRLF for SSH PTY output.
func RenderPrompt(scopeLabel, cmd string, lvl risk.Level) string {
	if lvl == risk.High {
		return fmt.Sprintf(
			"AiT: [%s] HIGH-RISK\n  %s\n  [y] run anyway\n  [n] skip\nChoice [y/N]: ",
			scopeLabel, cmd,
		)
	}
	return fmt.Sprintf(
		"AiT: [%s]\n  %s\n  [y] run once\n  [c] trust this command pattern\n  [s] trust safe commands this session\n  [n] skip\nChoice [y/c/s/N]: ",
		scopeLabel, cmd,
	)
}

func readLine(in io.Reader) (string, error) {
	if in == nil {
		return "", nil
	}
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", nil
	}
	return sc.Text(), nil
}
