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
	"strconv"
	"strings"
	"time"
)

const (
	// kmsgSource labels records that came from the kernel ring buffer.
	kmsgSource = "kernel"
	// kmsgPrefixSep ends the "level,seq,usec,flags" prefix of a kmsg line.
	kmsgPrefixSep = ';'
	// kmsgFieldSep separates the fields inside that prefix.
	kmsgFieldSep = ","
	// kmsgFacilityMod strips the syslog facility from the level field, leaving
	// the severity.
	kmsgFacilityMod = 8
	// kmsgTimeIndex is the field holding microseconds since boot.
	kmsgTimeIndex = 2
)

// ParseKmsgLine parses one /dev/kmsg record, formatted as
// "level,sequence,microseconds-since-boot,flags;message".
//
// boot is the wall-clock time the machine started, used to turn the
// monotonic-since-boot timestamp into an absolute one. Pass the zero time when
// boot time is unknown; the record then carries no timestamp, which time
// filters treat as unknown rather than as 1970.
//
// The second result is false only for a blank line. Lines without the expected
// prefix are still returned, with the whole line as the message.
func ParseKmsgLine(line string, boot time.Time) (Record, bool) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return Record{}, false
	}

	rec := Record{Raw: line, Message: line, Level: LevelUnset, Source: kmsgSource, Unit: kmsgSource}
	semi := strings.IndexByte(line, kmsgPrefixSep)
	if semi < 0 {
		return rec, true
	}
	rec.Message = line[semi+1:]

	fields := strings.Split(line[:semi], kmsgFieldSep)
	if n, err := strconv.Atoi(fields[0]); err == nil {
		rec.Level = FromSyslogPriority(n % kmsgFacilityMod)
	}
	if len(fields) > kmsgTimeIndex && !boot.IsZero() {
		rec.Time = kmsgTime(boot, fields[kmsgTimeIndex])
	}
	return rec, true
}

// kmsgTime converts microseconds since boot to wall-clock time, returning the
// zero time when the field is not a number.
func kmsgTime(boot time.Time, field string) time.Time {
	usec, err := strconv.ParseInt(field, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return boot.Add(time.Duration(usec) * time.Microsecond)
}
