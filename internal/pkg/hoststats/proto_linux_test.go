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
	"testing"

	"github.com/shirou/gopsutil/v4/net"
)

func TestApplyLinuxProtoStat(t *testing.T) {
	var raw protoRaw
	applyLinuxProtoCounters(&raw, []net.ProtoCountersStat{
		{Protocol: "tcp", Stats: map[string]int64{
			"InSegs": 10, "OutSegs": 20, "RetransSegs": 1, "CurrEstab": 3,
			"ActiveOpens": 4, "PassiveOpens": 5,
		}},
		{Protocol: "udp", Stats: map[string]int64{
			"InDatagrams": 7, "OutDatagrams": 8, "NoPorts": 2, "InErrors": 1,
		}},
		{Protocol: "icmp", Stats: map[string]int64{
			"InMsgs": 9, "OutMsgs": 11, "InErrors": 2, "OutErrors": 3,
		}},
	})
	if raw.tcpIn != 10 || raw.tcpOut != 20 || raw.tcpRetrans != 1 || raw.tcpEstab != 3 || raw.tcpOpens != 9 {
		t.Fatalf("tcp: %+v", raw)
	}
	if raw.udpIn != 7 || raw.udpOut != 8 || raw.udpNoPort != 2 || raw.udpInErr != 1 {
		t.Fatalf("udp: %+v", raw)
	}
	if raw.icmpIn != 9 || raw.icmpOut != 11 || raw.icmpInErr != 2 || raw.icmpOutErr != 3 {
		t.Fatalf("icmp: %+v", raw)
	}
}
