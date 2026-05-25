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

// Trust captures per-REPL exec-confirmation state: the current Policy,
// whether the user already opted into session-wide trust for low-risk commands,
// and a set of first-token patterns the user has chosen to auto-approve.
type Trust struct {
	// Policy controls whether each command needs confirmation or one approval lasts the session.
	Policy Policy
	// SessionTrusted is true once the user picked "s"; low-risk commands skip the prompt thereafter.
	SessionTrusted bool
	// TrustedPatterns holds first-token keys (e.g. "hostname") trusted via the "c" choice.
	TrustedPatterns map[string]struct{}
}

// IsPatternTrusted reports whether key has been added to the trust set.
// Returns false for empty key or nil receiver state.
func (t *Trust) IsPatternTrusted(key string) bool {
	if t == nil || key == "" || t.TrustedPatterns == nil {
		return false
	}
	_, ok := t.TrustedPatterns[key]
	return ok
}

// AddTrustedPattern marks key as trusted; subsequent low-risk commands matching
// key skip the confirmation prompt within the session. No-op for empty key.
func (t *Trust) AddTrustedPattern(key string) {
	if t == nil || key == "" {
		return
	}
	if t.TrustedPatterns == nil {
		t.TrustedPatterns = make(map[string]struct{})
	}
	t.TrustedPatterns[key] = struct{}{}
}

// Clear resets SessionTrusted and TrustedPatterns; Policy is preserved.
// Used on AI REPL exit to drop trust accrued during the session.
func (t *Trust) Clear() {
	if t == nil {
		return
	}
	t.SessionTrusted = false
	t.TrustedPatterns = nil
}

// ApplyResult records the trust implications of a Result.
// TrustSession sets SessionTrusted and switches Policy to PolicySessionOnce;
// TrustPattern adds patternKey to the trusted-pattern set.
func (t *Trust) ApplyResult(res Result, patternKey string) {
	if t == nil {
		return
	}
	if res.TrustSession {
		t.SessionTrusted = true
		t.Policy = PolicySessionOnce
	}
	if res.TrustPattern {
		t.AddTrustedPattern(patternKey)
	}
}
