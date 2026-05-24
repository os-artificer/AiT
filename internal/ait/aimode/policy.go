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
	"bufio"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// confirmExec returns whether the user approved running cmd under opt's scope and policy.
func confirmExec(opt Options, st *sessionState, cmdLabel, patternKey string, risk RiskLevel) (bool, error) {
	return confirmExecRisk(opt, st, cmdLabel, patternKey, risk)
}

// confirmExecRisk confirms execution using an explicit risk level and session trust state.
func confirmExecRisk(opt Options, st *sessionState, cmdLabel, patternKey string, risk RiskLevel) (bool, error) {
	if risk == RiskInteractive {
		return false, nil
	}

	if risk != RiskHigh && st.isPatternTrusted(patternKey) {
		return true, nil
	}
	if risk != RiskHigh && opt.ExecPolicy == ExecPolicySessionOnce && st.sessionTrusted {
		return true, nil
	}

	allowPattern := risk != RiskHigh
	allowSession := risk != RiskHigh
	prompt := confirmPromptRisk(opt, cmdLabel, risk)

	if opt.ConfirmRead != nil {
		res, err := opt.ConfirmRead(prompt)
		if err != nil {
			return false, err
		}
		if !res.Run {
			return false, nil
		}
		if res.TrustSession {
			st.sessionTrusted = true
			st.policy = ExecPolicySessionOnce
		}
		if res.TrustPattern {
			st.addTrustedPattern(patternKey)
		}
		return true, nil
	}

	aimWrite(opt.Out, prompt)
	line, err := readConfirmLine(opt.In)
	if err != nil {
		return false, err
	}

	res := parseConfirmChoice(line, allowPattern, allowSession)
	if !res.Run {
		return false, nil
	}
	if res.TrustSession {
		st.sessionTrusted = true
		st.policy = ExecPolicySessionOnce
	}
	if res.TrustPattern {
		st.addTrustedPattern(patternKey)
	}
	return true, nil
}

func scopeLabel(opt Options) string {
	if opt.Scope == ai.ScopeRemote {
		if opt.HostLabel != "" {
			return opt.HostLabel
		}
		return "remote"
	}
	return "local"
}

func readConfirmLine(in io.Reader) (string, error) {
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", nil
	}
	return sc.Text(), nil
}
