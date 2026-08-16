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

package logsrc

import "time"

// Backend selects which resolution layer to force.
type Backend string

const (
	// BackendAuto tries native then builtin layers in chain order.
	BackendAuto Backend = "auto"
	// BackendNative forces native-tool layers only.
	BackendNative Backend = "native"
	// BackendBuiltin forces pure-Go builtin layers only.
	BackendBuiltin Backend = "builtin"
)

// DefaultLimit is the default last-N matching records when -n is omitted.
const DefaultLimit = 1000

// Record is one normalized log entry from any backend.
type Record struct {
	Time   time.Time
	Host   string
	Source string
	Unit   string
	Level  Level
	PID    int
	// Process is the emitting process name when known (basename of
	// processImagePath, SYSLOG_IDENTIFIER, or syslog tag).
	Process string
	Message string
	Raw     string
}

// Query describes a log query. Filtering is always applied in Go; backends
// may narrow time/unit/level for performance but must not apply Limit.
type Query struct {
	Keywords   []string
	Regex      string
	Expr       string // raw --expr; when set, replaces Keywords/Regex text match
	IgnoreCase bool
	Since      time.Time
	Until      time.Time
	MinLevel   Level
	Unit       string
	Source     string
	Limit      int
	Head       bool
	Follow     bool
	Backend    Backend
	UTC        bool
	Raw        bool
	Quiet      bool
}

// EffectiveLimit returns Limit, or DefaultLimit when Limit is not positive.
// The CLI rejects non-positive -n values before building a Query; this
// fallback only guards programmatic callers that leave Limit unset.
func (q Query) EffectiveLimit() int {
	if q.Limit > 0 {
		return q.Limit
	}
	return DefaultLimit
}
