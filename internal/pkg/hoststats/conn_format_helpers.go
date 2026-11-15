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

package hoststats

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

var connTableFmt = fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds",
	connColProto, connColState, connColAddr, connColAddr, connColPID, connColProcess, connColAge, connColCreated)

const (
	connColProto   = 5
	connColState   = 12
	connColAddr    = 24
	connColPID     = 7
	connColProcess = 18
	connColAge     = 10
	connColCreated = 19
	connDashHuman  = "-"
	connCreatedFmt = "2006-01-02 15:04:05"
)

// FormatConnTable returns a human-readable connection table including the header.
func FormatConnTable(conns []ConnStat) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(connTableFmt,
		"PROTO", "STATE", "LOCAL", "REMOTE", "PID", "PROCESS", "AGE", "CREATED",
	))
	b.WriteByte('\n')
	for _, c := range conns {
		fmt.Fprintf(&b, connTableFmt+"\n",
			displayOrDash(c.Proto),
			displayOrDash(c.Status),
			formatConnEndpoint(c.LocalIP, c.LocalPort),
			formatConnRemote(c),
			formatConnPID(c.PID),
			formatConnProcess(c.Process),
			FormatConnAge(c.AgeSeconds),
			FormatConnCreated(c.CreatedAt),
		)
	}
	return b.String()
}

// WriteConnectionsJSON writes {"time":..., "connections":[...]} to w.
func WriteConnectionsJSON(w io.Writer, at time.Time, conns []ConnStat) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{
		"time":        at,
		"connections": conns,
	})
}

// FormatConnAge renders AgeSeconds as a compact duration, or "-" when nil.
func FormatConnAge(age *int64) string {
	if age == nil {
		return connDashHuman
	}
	sec := *age
	if sec < 0 {
		sec = 0
	}
	const (
		minute = int64(60)
		hour   = 60 * minute
		day    = 24 * hour
	)
	switch {
	case sec < minute:
		return fmt.Sprintf("%ds", sec)
	case sec < hour:
		m := sec / minute
		s := sec % minute
		if s == 0 {
			return fmt.Sprintf("%dm", m)
		}
		return fmt.Sprintf("%dm%ds", m, s)
	case sec < day:
		h := sec / hour
		m := (sec % hour) / minute
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	default:
		d := sec / day
		h := (sec % day) / hour
		if h == 0 {
			return fmt.Sprintf("%dd", d)
		}
		return fmt.Sprintf("%dd%dh", d, h)
	}
}

// FormatConnCreated renders CreatedAt as local wall time, or "-" when nil.
func FormatConnCreated(t *time.Time) string {
	if t == nil {
		return connDashHuman
	}
	return t.Local().Format(connCreatedFmt)
}

func formatConnEndpoint(ip string, port uint32) string {
	if ip == "" && port == 0 {
		return connDashHuman
	}
	var raw string
	switch {
	case ip == "":
		raw = fmt.Sprintf(":%d", port)
	case strings.Contains(ip, ":"):
		raw = fmt.Sprintf("[%s]:%d", ip, port)
	default:
		raw = fmt.Sprintf("%s:%d", ip, port)
	}
	return TruncateName(raw, connColAddr)
}

func formatConnRemote(c ConnStat) string {
	if c.Status == ConnStatusListen || (c.RemoteIP == "" && c.RemotePort == 0) {
		return connDashHuman
	}
	return formatConnEndpoint(c.RemoteIP, c.RemotePort)
}

func formatConnPID(pid int32) string {
	if pid <= 0 {
		return connDashHuman
	}
	return fmt.Sprintf("%d", pid)
}

func formatConnProcess(name string) string {
	if name == "" {
		return connDashHuman
	}
	return TruncateName(name, connColProcess)
}

func displayOrDash(s string) string {
	if s == "" {
		return connDashHuman
	}
	return s
}
