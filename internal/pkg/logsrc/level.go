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
	"fmt"
	"strconv"
	"strings"
)

// Level is a syslog-compatible severity (0=Emergency … 7=Debug).
type Level int

// Severities in syslog order, most severe first. LevelUnset marks a record
// whose backend did not report a severity; it is never filtered out by
// --priority, because the record might be of any severity.
const (
	LevelEmergency Level = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInfo
	LevelDebug

	LevelUnset Level = -1
)

var levelNames = []string{
	"emerg", "alert", "crit", "err", "warning", "notice", "info", "debug",
}

// String returns the short syslog name for lv.
func (lv Level) String() string {
	if lv < LevelEmergency || lv > LevelDebug {
		return "unknown"
	}
	return levelNames[lv]
}

// ParseLevel parses a priority name or numeric string into a Level.
// Accepts syslog names (emerg…debug), aliases (error, warn, critical),
// macOS names (default, fault), and Windows levels (1–5 as critical…verbose).
func ParseLevel(s string) (Level, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return LevelUnset, ErrBadLevel
	}
	if n, err := strconv.Atoi(s); err == nil {
		if n >= 0 && n <= 7 {
			return Level(n), nil
		}
		// Windows Event Level 1=Critical … 5=Verbose
		if n >= 1 && n <= 5 {
			return windowsLevelToSyslog(n), nil
		}
		return LevelUnset, fmt.Errorf("%w: %s", ErrBadLevel, s)
	}
	switch s {
	case "emerg", "emergency", "0":
		return LevelEmergency, nil
	case "alert", "1":
		return LevelAlert, nil
	case "crit", "critical", "fatal", "2":
		return LevelCritical, nil
	case "err", "error", "3":
		return LevelError, nil
	case "warning", "warn", "4":
		return LevelWarning, nil
	case "notice", "5":
		return LevelNotice, nil
	case "info", "information", "default", "6":
		return LevelInfo, nil
	case "debug", "verbose", "7":
		return LevelDebug, nil
	case "fault":
		return LevelCritical, nil
	default:
		return LevelUnset, fmt.Errorf("%w: %s", ErrBadLevel, s)
	}
}

// FromSyslogPriority maps a syslog priority number (0–7) to Level.
func FromSyslogPriority(n int) Level {
	if n < 0 || n > 7 {
		return LevelInfo
	}
	return Level(n)
}

// FromMacOS maps a macOS unified-log level string to Level.
func FromMacOS(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fault":
		return LevelCritical
	case "error":
		return LevelError
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	default:
		return LevelInfo
	}
}

// FromWindowsLevel maps a Windows Event Level (1–5) to Level.
func FromWindowsLevel(n int) Level {
	return windowsLevelToSyslog(n)
}

func windowsLevelToSyslog(n int) Level {
	switch n {
	case 1:
		return LevelCritical
	case 2:
		return LevelError
	case 3:
		return LevelWarning
	case 4:
		return LevelInfo
	case 5:
		return LevelDebug
	default:
		return LevelInfo
	}
}
