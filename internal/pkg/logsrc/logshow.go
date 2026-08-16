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

import (
	"encoding/json"
	"strings"
	"time"
)

// LogShowWindows are the --last windows tried in order, narrowest first.
// `log show` reads forward only and cannot be reversed, so a last-N query walks
// these windows until it has enough matches instead of scanning the whole store.
// The first window is short so unfiltered last-N (which must scan to EOF) stays
// responsive on a busy unified log.
var LogShowWindows = []time.Duration{
	5 * time.Minute,
	time.Hour,
	6 * time.Hour,
	24 * time.Hour,
	7 * 24 * time.Hour,
}

// LogShowWindowsFiltered is used when the query has keywords or a regex.
// Wider unified-log windows stay slow even with --predicate, so filtered
// queries stop at 1h unless the user passes an explicit --since.
var LogShowWindowsFiltered = []time.Duration{
	5 * time.Minute,
	15 * time.Minute,
	time.Hour,
}

const (
	// LogShowStyleNDJSON is the structured output, preferred because it parses
	// unambiguously.
	LogShowStyleNDJSON = "ndjson"
	// LogShowStyleSyslog is the human-readable output, used on macOS releases
	// that predate --style ndjson.
	LogShowStyleSyslog = "syslog"

	// logShowTimeLayout is the timestamp format `log show` accepts for
	// --start and --end.
	logShowTimeLayout = "2006-01-02 15:04:05"
	// logShowSyslogTimeLayout is the timestamp format the syslog style prints.
	logShowSyslogTimeLayout = "2006-01-02 15:04:05.000000-0700"

	// logShowSyslogMinFields is the smallest field count that can hold a
	// two-field timestamp plus a message.
	logShowSyslogMinFields = 3
)

// BuildLogShowArgs builds argv for `log show` (without the binary name).
// window bounds the query with --last when q.Since is zero; a zero window
// leaves the query unbounded. Following uses BuildLogStreamArgs instead.
func BuildLogShowArgs(q Query, window time.Duration, ndjson bool) []string {
	style := LogShowStyleSyslog
	if ndjson {
		style = LogShowStyleNDJSON
	}
	args := []string{"show", "--style", style}

	if !q.Since.IsZero() {
		args = append(args, "--start", q.Since.Format(logShowTimeLayout))
	} else if window > 0 {
		args = append(args, "--last", FormatWindow(window))
	}
	if !q.Until.IsZero() {
		args = append(args, "--end", q.Until.Format(logShowTimeLayout))
	}
	args = appendLogShowVerbosity(args, q)
	return appendLogShowPredicate(args, q)
}

// BuildLogStreamArgs builds argv for `log stream`, used for --follow.
// Streaming always uses the structured style.
func BuildLogStreamArgs(q Query) []string {
	args := []string{"stream", "--style", LogShowStyleNDJSON}
	args = appendLogShowVerbosity(args, q)
	return appendLogShowPredicate(args, q)
}

// appendLogShowVerbosity adds --info/--debug when the priority floor needs those
// macOS message types. There is no `log show --level`; exact --priority filtering
// still happens in Go.
func appendLogShowVerbosity(args []string, q Query) []string {
	if q.MinLevel == LevelUnset {
		return args
	}
	if q.MinLevel >= LevelInfo {
		args = append(args, "--info")
	}
	if q.MinLevel >= LevelDebug {
		args = append(args, "--debug")
	}
	return args
}

type logShowNDJSON struct {
	Timestamp   string `json:"timestamp"`
	EventType   string `json:"eventType"`
	EventMsg    string `json:"eventMessage"`
	ProcessID   int64  `json:"processID"`
	ProcessImg  string `json:"processImagePath"`
	Subsystem   string `json:"subsystem"`
	Category    string `json:"category"`
	MessageType string `json:"messageType"`
}

// ParseLogShowNDJSON parses one `log show --style ndjson` line.
func ParseLogShowNDJSON(line []byte) (Record, error) {
	var j logShowNDJSON
	if err := json.Unmarshal(line, &j); err != nil {
		return Record{}, err
	}
	rec := Record{
		Raw:     string(line),
		Message: j.EventMsg,
		PID:     int(j.ProcessID),
		Process: processBaseName(j.ProcessImg),
		Unit:    firstNonEmpty(j.Subsystem, j.ProcessImg, j.Category),
		Level:   FromMacOS(j.MessageType),
		Source:  SourceNameSystem,
	}
	rec.Time = parseLogShowTime(j.Timestamp)
	return rec, nil
}

// ParseLogShowSyslog parses a `log show --style syslog` line (secondary).
func ParseLogShowSyslog(line string) (Record, bool) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return Record{}, false
	}
	// "Timestamp process[pid]: message" — the style carries no level, so the
	// record stays LevelUnset and passes any --priority floor.
	rec := Record{Raw: line, Message: line, Level: LevelUnset, Source: SourceNameSystem}
	fields := strings.Fields(line)
	if len(fields) < logShowSyslogMinFields {
		return rec, true
	}

	stamp := fields[0] + " " + fields[1]
	t, err := time.ParseInLocation(logShowSyslogTimeLayout, stamp, time.Local)
	if err != nil {
		return rec, true
	}
	rec.Time = t
	rec.Message = strings.TrimPrefix(line, stamp+" ")
	applySyslogProcess(rec.Message, &rec)
	return rec, true
}

// parseLogShowTime parses an ndjson timestamp, accepting both the RFC 3339 form
// and the space-separated form older releases emit. Returns the zero time when
// the field is absent or unparsable, which time filters treat as unknown.
func parseLogShowTime(stamp string) time.Time {
	if stamp == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, stamp); err == nil {
		return t
	}
	if t, err := time.Parse(logShowSyslogTimeLayout, stamp); err == nil {
		return t
	}
	return time.Time{}
}
