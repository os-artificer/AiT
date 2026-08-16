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
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BuildJournalctlArgs builds argv for journalctl (without the binary name).
// When reverse is true, --reverse is added. Record caps are never passed
// when content filtering is active (limit applied in Go).
func BuildJournalctlArgs(q Query, reverse bool, kernel bool) []string {
	args := []string{"-o", "json", "--no-pager"}
	if reverse && !q.Follow {
		args = append(args, "--reverse")
	}
	if kernel {
		args = append(args, "-k")
	}
	if q.Follow {
		args = append(args, "-f")
	}
	if !q.Since.IsZero() {
		args = append(args, "--since", q.Since.Format("2006-01-02 15:04:05"))
	}
	if !q.Until.IsZero() {
		args = append(args, "--until", q.Until.Format("2006-01-02 15:04:05"))
	}
	if q.MinLevel != LevelUnset {
		args = append(args, "-p", strconv.Itoa(int(q.MinLevel)))
	}
	if q.Unit != "" {
		args = append(args, "-u", q.Unit)
	}
	return appendJournalGrep(args, q)
}

type journalJSON struct {
	Realtime string `json:"__REALTIME_TIMESTAMP"`
	Message  any    `json:"MESSAGE"`
	Priority string `json:"PRIORITY"`
	Hostname string `json:"_HOSTNAME"`
	Unit     string `json:"_SYSTEMD_UNIT"`
	SyslogID string `json:"SYSLOG_IDENTIFIER"`
	PID      string `json:"_PID"`
}

// ParseJournalJSON parses one journalctl -o json line into a Record.
func ParseJournalJSON(line []byte) (Record, error) {
	var j journalJSON
	if err := json.Unmarshal(line, &j); err != nil {
		return Record{}, err
	}
	rec := Record{
		Host:    j.Hostname,
		Unit:    firstNonEmpty(j.Unit, j.SyslogID),
		Process: firstNonEmpty(j.SyslogID, strings.TrimSuffix(j.Unit, ".service")),
		Raw:     string(line),
		Message: stringifyMessage(j.Message),
		Level:   LevelInfo,
		Source:  SourceNameSystem,
	}
	if j.Priority != "" {
		if n, err := strconv.Atoi(j.Priority); err == nil {
			rec.Level = FromSyslogPriority(n)
		}
	}
	if j.PID != "" {
		if n, err := strconv.Atoi(j.PID); err == nil {
			rec.PID = n
		}
	}
	if j.Realtime != "" {
		if us, err := strconv.ParseInt(j.Realtime, 10, 64); err == nil {
			rec.Time = time.UnixMicro(us)
		}
	}
	return rec, nil
}

// ParseJournalShortISO parses a journalctl short-iso line (secondary parser).
func ParseJournalShortISO(line string) (Record, bool) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return Record{}, false
	}
	// 2024-01-02T15:04:05+0000 host unit[pid]: message
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return Record{Raw: line, Message: line, Level: LevelInfo}, true
	}
	t, err := time.Parse("2006-01-02T15:04:05-0700", parts[0])
	if err != nil {
		t, err = time.Parse(time.RFC3339, parts[0])
	}
	rec := Record{
		Raw:     line,
		Host:    parts[1],
		Message: parts[3],
		Level:   LevelInfo,
		Source:  SourceNameSystem,
	}
	if err == nil {
		rec.Time = t
	}
	unit := parts[2]
	if i := strings.IndexByte(unit, '['); i >= 0 {
		rec.Unit = unit[:i]
		rec.Process = rec.Unit
		pidStr := strings.TrimSuffix(unit[i+1:], "]:")
		pidStr = strings.TrimSuffix(pidStr, "]")
		if n, err := strconv.Atoi(pidStr); err == nil {
			rec.PID = n
		}
	} else {
		rec.Unit = strings.TrimSuffix(unit, ":")
		rec.Process = rec.Unit
	}
	return rec, true
}

func stringifyMessage(m any) string {
	switch v := m.(type) {
	case string:
		return v
	case []any:
		b := make([]byte, 0, len(v))
		for _, x := range v {
			switch n := x.(type) {
			case float64:
				b = append(b, byte(n))
			}
		}
		return string(b)
	default:
		return fmt.Sprint(m)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
