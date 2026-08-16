//go:build darwin

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
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
)

const (
	cmdNetstat  = "netstat"
	cmdARP      = "arp"
	flagStats   = "-s"
	flagNumeric = "-an"

	darwinARPSep        = " at "
	darwinARPIncomplete = "(incomplete)"
	tcpStateEstablished = "established"

	darwinTCPSent     = "packet sent"
	darwinTCPRecv     = "packet received"
	darwinTCPConnReq  = "connection request"
	darwinTCPConnAcc  = "connection accept"
	darwinTCPEstab    = "connection established"
	darwinUDPDgramIn  = "datagram received"
	darwinUDPDgramOut = "datagram output"
)

func sampleProtoRaw(ctx context.Context) (protoRaw, error) {
	var raw protoRaw
	if out := runCmdOutput(ctx, cmdNetstat, flagStats); out != nil {
		fillDarwinNetstat(&raw, out)
	}
	raw.tcpEstab = countDarwinTCPEstab(ctx)
	ents, incomp := countDarwinARP(ctx)
	raw.arpEnts = ents
	raw.arpIncomp = incomp
	return raw, nil
}

func fillDarwinNetstat(raw *protoRaw, data []byte) {
	section := ""
	eachNonEmptyLine(data, func(s string) {
		lower := strings.ToLower(s)
		if next, ok := darwinNetstatSection(lower); ok {
			section = next
			return
		}
		n, rest, ok := leadingUint(s)
		if !ok {
			return
		}
		switch section {
		case protoNameTCP:
			applyDarwinTCP(raw, n, rest)
		case protoNameUDP:
			applyDarwinUDP(raw, n, rest)
		case protoNameICMP:
			applyDarwinICMP(raw, n, rest)
		}
	})
}

func darwinNetstatSection(lower string) (string, bool) {
	switch {
	case strings.HasPrefix(lower, protoNameTCP+":"):
		return protoNameTCP, true
	case strings.HasPrefix(lower, protoNameUDP+":"):
		return protoNameUDP, true
	case strings.HasPrefix(lower, protoNameICMP+":"):
		return protoNameICMP, true
	case strings.HasSuffix(lower, ":") && !strings.Contains(lower, " "):
		return "", true
	}
	return "", false
}

func applyDarwinTCP(raw *protoRaw, n uint64, rest string) {
	switch {
	case darwinExactMetric(rest, darwinTCPSent):
		raw.tcpOut = n
	case darwinExactMetric(rest, darwinTCPRecv):
		raw.tcpIn = n
	case strings.Contains(rest, "retransmit"):
		raw.tcpRetrans += n
	case darwinExactMetric(rest, darwinTCPConnReq):
		raw.tcpOpens += n
	case darwinExactMetric(rest, darwinTCPConnAcc):
		raw.tcpOpens += n
	case darwinMetricPrefix(rest, darwinTCPEstab):
		raw.tcpEstab = n
	}
}

func applyDarwinUDP(raw *protoRaw, n uint64, rest string) {
	switch {
	case darwinExactMetric(rest, darwinUDPDgramIn) || darwinExactMetric(rest, darwinTCPRecv):
		raw.udpIn = n
	case darwinExactMetric(rest, darwinUDPDgramOut) || darwinExactMetric(rest, darwinTCPSent):
		raw.udpOut = n
	case strings.Contains(rest, "dropped due to no socket"):
		raw.udpNoPort = n
	case strings.Contains(rest, "with bad checksum") || strings.Contains(rest, "input errors"):
		raw.udpInErr += n
	}
}

func applyDarwinICMP(raw *protoRaw, n uint64, rest string) {
	switch {
	case strings.HasPrefix(rest, "messages received") || strings.HasPrefix(rest, "calls to icmp_input"):
		raw.icmpIn = n
	case strings.HasPrefix(rest, "messages sent") || strings.HasPrefix(rest, "messages output"):
		raw.icmpOut = n
	case strings.HasPrefix(rest, "input histogram") || strings.HasPrefix(rest, "output histogram"):
		return
	case strings.Contains(rest, "bad") || strings.Contains(rest, "error"):
		applyDarwinICMPErr(raw, n, rest)
	}
}

func applyDarwinICMPErr(raw *protoRaw, n uint64, rest string) {
	if strings.Contains(rest, "out") || strings.Contains(rest, "sent") {
		raw.icmpOutErr += n
		return
	}
	raw.icmpInErr += n
}

func darwinExactMetric(rest, singular string) bool {
	if rest == singular {
		return true
	}
	words := strings.Fields(singular)
	if len(words) == 0 {
		return false
	}
	first := append([]string{words[0] + "s"}, words[1:]...)
	last := append([]string{}, words...)
	last[len(last)-1] += "s"
	return rest == strings.Join(first, " ") || rest == strings.Join(last, " ")
}

func darwinMetricPrefix(rest, singular string) bool {
	if strings.HasPrefix(rest, singular) {
		return true
	}
	words := strings.Fields(singular)
	if len(words) == 0 {
		return false
	}
	first := append([]string{words[0] + "s"}, words[1:]...)
	last := append([]string{}, words...)
	last[len(last)-1] += "s"
	return strings.HasPrefix(rest, strings.Join(first, " ")) || strings.HasPrefix(rest, strings.Join(last, " "))
}

func leadingUint(s string) (uint64, string, bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, "", false
	}
	v, err := strconv.ParseUint(s[:i], 10, 64)
	if err != nil {
		return 0, "", false
	}
	return v, strings.ToLower(strings.TrimSpace(s[i:])), true
}

func countDarwinTCPEstab(ctx context.Context) uint64 {
	return countEstablishedTCP(runCmdOutput(ctx, cmdNetstat, flagNumeric))
}

func countEstablishedTCP(data []byte) uint64 {
	var n uint64
	eachNonEmptyLine(data, func(s string) {
		s = strings.ToLower(s)
		if strings.HasPrefix(s, protoNameTCP) && strings.Contains(s, tcpStateEstablished) {
			n++
		}
	})
	return n
}

func countDarwinARP(ctx context.Context) (entries, incomplete uint64) {
	eachNonEmptyLine(runCmdOutput(ctx, cmdARP, flagNumeric), func(s string) {
		if !strings.Contains(s, darwinARPSep) {
			return
		}
		entries++
		if strings.Contains(s, darwinARPIncomplete) {
			incomplete++
		}
	})
	return entries, incomplete
}

func runCmdOutput(ctx context.Context, name string, args ...string) []byte {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return nil
	}
	return out
}

func eachNonEmptyLine(data []byte, fn func(string)) {
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		s := strings.TrimSpace(string(line))
		if s == "" {
			continue
		}
		fn(s)
	}
}
