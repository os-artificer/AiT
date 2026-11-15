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
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// KnownConnStatuses lists canonical Status values accepted by ConnQuery / --state.
var KnownConnStatuses = []string{
	ConnStatusListen,
	ConnStatusEstablished,
	ConnStatusSynSent,
	ConnStatusSynRecv,
	ConnStatusFinWait1,
	ConnStatusFinWait2,
	ConnStatusTimeWait,
	ConnStatusClose,
	ConnStatusCloseWait,
	ConnStatusLastAck,
	ConnStatusClosing,
	ConnStatusNone,
	ConnStatusDelete,
}

// connStatusCanonical maps OS/gopsutil spellings and aliases to canonical names.
var connStatusCanonical = map[string]string{
	ConnStatusListen:      ConnStatusListen,
	ConnStatusEstablished: ConnStatusEstablished,
	ConnStatusSynSent:     ConnStatusSynSent,
	ConnStatusSynRecv:     ConnStatusSynRecv,
	"SYN_RECEIVED":        ConnStatusSynRecv,
	"SYN_RCVD":            ConnStatusSynRecv,
	ConnStatusFinWait1:    ConnStatusFinWait1,
	"FIN_WAIT_1":          ConnStatusFinWait1,
	ConnStatusFinWait2:    ConnStatusFinWait2,
	"FIN_WAIT_2":          ConnStatusFinWait2,
	ConnStatusTimeWait:    ConnStatusTimeWait,
	ConnStatusClose:       ConnStatusClose,
	"CLOSED":              ConnStatusClose,
	ConnStatusCloseWait:   ConnStatusCloseWait,
	ConnStatusLastAck:     ConnStatusLastAck,
	ConnStatusClosing:     ConnStatusClosing,
	ConnStatusNone:        ConnStatusNone,
	ConnStatusDelete:      ConnStatusDelete,
}

// NormalizeConnQuery applies defaults and clamps Limit.
// Returns an error when Proto or States contain invalid values, or Limit is negative.
func NormalizeConnQuery(q ConnQuery) (ConnQuery, error) {
	proto := strings.ToLower(strings.TrimSpace(q.Proto))
	if proto == "" {
		proto = ConnProtoTCP
	}
	switch proto {
	case ConnProtoTCP, ConnProtoUDP, ConnProtoAll:
		q.Proto = proto
	default:
		return ConnQuery{}, fmt.Errorf("invalid connection proto %q", q.Proto)
	}

	if len(q.States) > 0 {
		normalized, err := normalizeConnStates(q.States)
		if err != nil {
			return ConnQuery{}, err
		}
		q.States = normalized
	}

	if q.Limit < 0 {
		return ConnQuery{}, fmt.Errorf("invalid connection limit %d", q.Limit)
	}
	if q.Limit == 0 {
		q.Limit = DefaultConnLimit
	}
	if q.Limit > MaxConnLimit {
		q.Limit = MaxConnLimit
	}
	return q, nil
}

// ListConnections returns filtered, sorted, limited local connections.
// Proto "all" maps to gopsutil kind "inet" (tcp+udp, no unix).
// When PID > 0, uses ConnectionsPidWithContext. Age enrichment is Linux-best-effort.
// Returns NormalizeConnQuery errors or a gopsutil connections error.
func ListConnections(ctx context.Context, q ConnQuery) ([]ConnStat, error) {
	q, err := NormalizeConnQuery(q)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	kind := gopsutilConnKind(q.Proto)
	raw, err := fetchConnections(ctx, kind, q.PID)
	if err != nil {
		return nil, err
	}

	out := make([]ConnStat, 0, len(raw))
	for _, s := range raw {
		c := connStatFromGopsutil(s)
		if !connMatchesQuery(c, q) {
			continue
		}
		out = append(out, c)
	}

	sortConnStats(out)
	if len(out) > q.Limit {
		out = out[:q.Limit]
	}

	enrichConnAge(out)
	attachConnProcessNames(ctx, out)
	return out, nil
}

func gopsutilConnKind(proto string) string {
	if proto == ConnProtoAll {
		return "inet"
	}
	return proto
}

func fetchConnections(ctx context.Context, kind string, pid int32) ([]net.ConnectionStat, error) {
	if pid > 0 {
		return net.ConnectionsPidWithContext(ctx, kind, pid)
	}
	return net.ConnectionsWithContext(ctx, kind)
}

func connStatFromGopsutil(s net.ConnectionStat) ConnStat {
	status := strings.ToUpper(strings.TrimSpace(s.Status))
	if canon, ok := canonicalConnStatus(status); ok {
		status = canon
	}
	return ConnStat{
		Proto:      sockTypeProto(s.Type),
		Family:     familyLabel(s.Family),
		Status:     status,
		LocalIP:    s.Laddr.IP,
		LocalPort:  s.Laddr.Port,
		RemoteIP:   s.Raddr.IP,
		RemotePort: s.Raddr.Port,
		PID:        s.Pid,
		Fd:         s.Fd,
	}
}

func sockTypeProto(t uint32) string {
	switch t {
	case syscall.SOCK_STREAM:
		return ConnProtoTCP
	case syscall.SOCK_DGRAM:
		return ConnProtoUDP
	default:
		return ""
	}
}

func familyLabel(f uint32) string {
	switch f {
	case syscall.AF_INET:
		return ConnFamilyIPv4
	case syscall.AF_INET6:
		return ConnFamilyIPv6
	default:
		return ""
	}
}

func connMatchesQuery(c ConnStat, q ConnQuery) bool {
	if q.PID > 0 && c.PID != q.PID {
		return false
	}
	if q.Port > 0 && int(c.LocalPort) != q.Port && int(c.RemotePort) != q.Port {
		return false
	}
	if len(q.States) == 0 {
		return true
	}
	for _, want := range q.States {
		if c.Status == want {
			return true
		}
	}
	return false
}

func normalizeConnStates(states []string) ([]string, error) {
	out := make([]string, 0, len(states))
	seen := map[string]struct{}{}
	for _, raw := range states {
		for _, part := range strings.Split(raw, ",") {
			s := strings.ToUpper(strings.TrimSpace(part))
			if s == "" {
				continue
			}
			canon, ok := canonicalConnStatus(s)
			if !ok {
				return nil, fmt.Errorf("invalid connection state %q", part)
			}
			if _, dup := seen[canon]; dup {
				continue
			}
			seen[canon] = struct{}{}
			out = append(out, canon)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty connection state filter")
	}
	return out, nil
}

// canonicalConnStatus maps a trimmed uppercase status to the canonical name.
// Returns false when s is not a known status or alias.
func canonicalConnStatus(s string) (string, bool) {
	canon, ok := connStatusCanonical[s]
	return canon, ok
}

func sortConnStats(conns []ConnStat) {
	sort.SliceStable(conns, func(i, j int) bool {
		a, b := conns[i], conns[j]
		if a.Proto != b.Proto {
			return a.Proto < b.Proto
		}
		if a.Status != b.Status {
			return a.Status < b.Status
		}
		if a.LocalIP != b.LocalIP {
			return a.LocalIP < b.LocalIP
		}
		if a.LocalPort != b.LocalPort {
			return a.LocalPort < b.LocalPort
		}
		if a.RemoteIP != b.RemoteIP {
			return a.RemoteIP < b.RemoteIP
		}
		if a.RemotePort != b.RemotePort {
			return a.RemotePort < b.RemotePort
		}
		return a.PID < b.PID
	})
}

func attachConnProcessNames(ctx context.Context, conns []ConnStat) {
	cache := map[int32]string{}
	for i := range conns {
		pid := conns[i].PID
		if pid <= 0 {
			continue
		}
		if name, ok := cache[pid]; ok {
			conns[i].Process = name
			continue
		}
		name := lookupProcessName(ctx, pid)
		cache[pid] = name
		conns[i].Process = name
	}
}

func lookupProcessName(ctx context.Context, pid int32) string {
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return ""
	}
	name, err := p.NameWithContext(ctx)
	if err == nil && name != "" {
		return name
	}
	cmdline, err := p.CmdlineSliceWithContext(ctx)
	if err != nil {
		return processDisplayName(name, nil)
	}
	return processDisplayName(name, cmdline)
}

// processDisplayName prefers a non-empty process name; otherwise uses cmdline[0] basename.
func processDisplayName(name string, cmdline []string) string {
	if name != "" {
		return name
	}
	if len(cmdline) == 0 || cmdline[0] == "" {
		return ""
	}
	return filepath.Base(cmdline[0])
}
