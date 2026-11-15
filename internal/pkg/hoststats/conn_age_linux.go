//go:build linux

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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	nfConntrackPath      = "/proc/net/nf_conntrack"
	procFDPathFmt        = "/proc/%d/fd/%d"
	conntrackScanBufInit = 64 * 1024
	conntrackScanBufMax  = 1024 * 1024
	deltaTimeKey         = "delta-time="
	srcKey               = "src="
	dstKey               = "dst="
	sportKey             = "sport="
	dportKey             = "dport="
)

type connTuple struct {
	proto        string
	lip, rip     string
	lport, rport uint32
}

// enrichConnAge sets CreatedAt and AgeSeconds from /proc/<pid>/fd/<fd> Ctim when possible.
// On failure, tries nf_conntrack delta-time (AgeSeconds + derived CreatedAt).
func enrichConnAge(conns []ConnStat) {
	if len(conns) == 0 {
		return
	}

	now := time.Now()
	enrichConnAgeWith(conns, now, nil)
	for i := range conns {
		if conns[i].CreatedAt == nil {
			enrichConnAgeWith(conns, now, loadConntrackAges())
			return
		}
	}
}

// enrichConnAgeWith applies fd then optional track enrichment (for tests with an injected map).
func enrichConnAgeWith(conns []ConnStat, now time.Time, track map[connTuple]int64) {
	for i := range conns {
		if applyFDConnAge(&conns[i], now) {
			continue
		}
		applyConntrackConnAge(&conns[i], now, track)
	}
}

func applyFDConnAge(c *ConnStat, now time.Time) bool {
	created := createdAtFromFD(c.PID, c.Fd)
	if created == nil || created.After(now) {
		return false
	}

	t := *created
	c.CreatedAt = &t
	c.AgeSeconds = ageSecondsSince(t, now)
	return true
}

func applyConntrackConnAge(c *ConnStat, now time.Time, track map[connTuple]int64) {
	if track == nil {
		return
	}
	age := ageFromConntrack(track, *c)
	if age == nil || *age < 0 {
		return
	}
	sec := *age
	c.AgeSeconds = &sec
	t := now.Add(-time.Duration(sec) * time.Second)
	c.CreatedAt = &t
}

// createdAtFromFD returns the fd create/mtime instant from /proc/<pid>/fd/<fd>.
// Returns nil when pid <= 0, Lstat fails, or the timestamp is zero.
func createdAtFromFD(pid int32, fd uint32) *time.Time {
	if pid <= 0 {
		return nil
	}

	path := fmt.Sprintf(procFDPathFmt, pid, fd)
	fi, err := os.Lstat(path)
	if err != nil {
		return nil
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nonZeroTime(fi.ModTime())
	}
	sec, nsec := st.Ctim.Unix()
	if sec == 0 && nsec == 0 {
		sec, nsec = st.Mtim.Unix()
	}
	return nonZeroTime(time.Unix(sec, nsec))
}

func nonZeroTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// ageSecondsSince returns whole seconds from created to now, or nil if invalid.
func ageSecondsSince(created, now time.Time) *int64 {
	if created.IsZero() || created.After(now) {
		return nil
	}
	sec := int64(now.Sub(created).Seconds())
	if sec < 0 {
		sec = 0
	}
	return &sec
}

func ageFromConntrack(track map[connTuple]int64, c ConnStat) *int64 {
	if len(track) == 0 {
		return nil
	}
	key := connTuple{
		proto: c.Proto,
		lip:   c.LocalIP,
		rip:   c.RemoteIP,
		lport: c.LocalPort,
		rport: c.RemotePort,
	}
	if age, ok := track[key]; ok {
		v := age
		return &v
	}
	key.lip, key.rip = key.rip, key.lip
	key.lport, key.rport = key.rport, key.lport
	if age, ok := track[key]; ok {
		v := age
		return &v
	}
	return nil
}

func loadConntrackAges() map[connTuple]int64 {
	f, err := os.Open(nfConntrackPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	out := map[connTuple]int64{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, conntrackScanBufInit), conntrackScanBufMax)
	for sc.Scan() {
		t, age, ok := parseConntrackLine(sc.Text())
		if !ok {
			continue
		}
		out[t] = age
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseConntrackLine extracts a 4-tuple and delta-time when present.
func parseConntrackLine(line string) (connTuple, int64, bool) {
	if !strings.Contains(line, deltaTimeKey) {
		return connTuple{}, 0, false
	}
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return connTuple{}, 0, false
	}

	proto := conntrackProto(fields)
	if proto != ConnProtoTCP && proto != ConnProtoUDP {
		return connTuple{}, 0, false
	}

	var src, dst string
	var sport, dport uint32
	var age int64
	haveAge := false
	for _, f := range fields {
		src = firstConntrackKV(src, f, srcKey)
		dst = firstConntrackKV(dst, f, dstKey)
		sport = firstConntrackUint32(sport, f, sportKey)
		dport = firstConntrackUint32(dport, f, dportKey)
		age, haveAge = firstConntrackDelta(age, haveAge, f)
	}
	if !haveAge || src == "" || dst == "" {
		return connTuple{}, 0, false
	}
	return connTuple{proto: proto, lip: src, rip: dst, lport: sport, rport: dport}, age, true
}

func conntrackProto(fields []string) string {
	proto := strings.ToLower(fields[0])
	if proto == ConnProtoTCP || proto == ConnProtoUDP {
		return proto
	}
	if len(fields) > 2 {
		return strings.ToLower(fields[2])
	}
	return proto
}

func firstConntrackKV(cur, f, key string) string {
	if cur != "" || !strings.HasPrefix(f, key) {
		return cur
	}
	return strings.TrimPrefix(f, key)
}

func firstConntrackUint32(cur uint32, f, key string) uint32 {
	if cur != 0 || !strings.HasPrefix(f, key) {
		return cur
	}
	return parseUint32Field(f, key)
}

func firstConntrackDelta(age int64, have bool, f string) (int64, bool) {
	if !strings.HasPrefix(f, deltaTimeKey) {
		return age, have
	}
	v, err := strconv.ParseInt(strings.TrimPrefix(f, deltaTimeKey), 10, 64)
	if err != nil {
		return age, have
	}
	return v, true
}

func parseUint32Field(f, prefix string) uint32 {
	v, err := strconv.ParseUint(strings.TrimPrefix(f, prefix), 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}
