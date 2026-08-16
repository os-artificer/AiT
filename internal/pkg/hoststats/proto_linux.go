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
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
)

const (
	procNetARPPath   = "/proc/net/arp"
	procARPCachePath = "/proc/net/stat/arp_cache"
	arpCompleteFlag  = 0x2
)

func sampleProtoRaw(ctx context.Context) (protoRaw, error) {
	var raw protoRaw
	stats, err := net.ProtoCountersWithContext(ctx, []string{protoNameTCP, protoNameUDP, protoNameICMP})
	if err == nil {
		applyLinuxProtoCounters(&raw, stats)
	}
	fillLinuxARP(&raw)
	return raw, nil
}

func applyLinuxProtoCounters(raw *protoRaw, stats []net.ProtoCountersStat) {
	for _, s := range stats {
		applyLinuxProtoStat(raw, s)
	}
}

func applyLinuxProtoStat(raw *protoRaw, s net.ProtoCountersStat) {
	switch strings.ToLower(s.Protocol) {
	case protoNameTCP:
		raw.tcpIn = protoMapGet(s.Stats, "InSegs")
		raw.tcpOut = protoMapGet(s.Stats, "OutSegs")
		raw.tcpRetrans = protoMapGet(s.Stats, "RetransSegs")
		raw.tcpEstab = protoMapGet(s.Stats, "CurrEstab")
		raw.tcpOpens = protoMapGet(s.Stats, "ActiveOpens") + protoMapGet(s.Stats, "PassiveOpens")
	case protoNameUDP:
		raw.udpIn = protoMapGet(s.Stats, "InDatagrams")
		raw.udpOut = protoMapGet(s.Stats, "OutDatagrams")
		raw.udpNoPort = protoMapGet(s.Stats, "NoPorts")
		raw.udpInErr = protoMapGet(s.Stats, "InErrors")
	case protoNameICMP:
		raw.icmpIn = protoMapGet(s.Stats, "InMsgs")
		raw.icmpOut = protoMapGet(s.Stats, "OutMsgs")
		raw.icmpInErr = protoMapGet(s.Stats, "InErrors")
		raw.icmpOutErr = protoMapGet(s.Stats, "OutErrors")
	}
}

func fillLinuxARP(raw *protoRaw) {
	ents, incomp := readProcNetARP(procNetARPPath)
	raw.arpEnts = ents
	raw.arpIncomp = incomp
	lookups, reqs, replies := readProcARPCache(procARPCachePath)
	raw.arpLookups = lookups
	raw.arpOut = reqs
	raw.arpIn = replies
}

func readProcNetARP(path string) (entries, incomplete uint64) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0, 0
	}
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 {
			continue
		}
		entries++
		flags, err := strconv.ParseUint(fields[2], 0, 64)
		if err != nil || flags&arpCompleteFlag == 0 {
			incomplete++
		}
	}
	return entries, incomplete
}

func readProcARPCache(path string) (lookups, reqs, replies uint64) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var headers []string
	for sc.Scan() {
		fields := arpCacheFields(sc.Text())
		if fields == nil {
			continue
		}
		if len(headers) == 0 {
			headers = fields
			continue
		}
		return parseARPCacheRow(headers, fields)
	}
	return 0, 0, 0
}

func arpCacheFields(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasSuffix(line, ":") {
		return nil
	}
	return strings.Fields(line)
}

func parseARPCacheRow(headers, fields []string) (lookups, reqs, replies uint64) {
	if len(fields) != len(headers) {
		return 0, 0, 0
	}
	idx := headerIndexMap(headers)
	lookups = parseU64Field(fields, idx, "lookups")
	reqs = parseU64Field(fields, idx, "req_sent", "requests")
	replies = parseU64Field(fields, idx, "replies_recv", "replies")
	return lookups, reqs, replies
}

func headerIndexMap(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.ToLower(h)] = i
	}
	return idx
}

func parseU64Field(fields []string, idx map[string]int, keys ...string) uint64 {
	for _, k := range keys {
		i, ok := idx[k]
		if !ok || i < 0 || i >= len(fields) {
			continue
		}
		v, err := strconv.ParseUint(fields[i], 10, 64)
		if err == nil {
			return v
		}
	}
	return 0
}
