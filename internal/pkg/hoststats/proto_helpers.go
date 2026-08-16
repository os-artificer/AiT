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

const (
	protoNameTCP  = "tcp"
	protoNameUDP  = "udp"
	protoNameICMP = "icmp"
)

// protoRaw holds absolute protocol counters used to derive per-second rates.
type protoRaw struct {
	tcpIn, tcpOut, tcpRetrans, tcpEstab, tcpOpens uint64
	udpIn, udpOut, udpNoPort, udpInErr            uint64
	icmpIn, icmpOut, icmpInErr, icmpOutErr        uint64
	arpIn, arpOut, arpEnts, arpIncomp, arpLookups uint64
}

func protoStatsFromRaw(cur, prev protoRaw, dt float64, havePrev bool) ProtoStats {
	out := ProtoStats{
		TCP: TCPStat{
			Estab: cur.tcpEstab,
			Opens: cur.tcpOpens,
		},
		UDP: UDPStat{
			NoPort: cur.udpNoPort,
			InErr:  cur.udpInErr,
		},
		ARP: ARPStat{
			Entries:    cur.arpEnts,
			Incomplete: cur.arpIncomp,
		},
		ICMP: ICMPStat{
			InErr:  cur.icmpInErr,
			OutErr: cur.icmpOutErr,
		},
	}
	if !havePrev || dt <= 0 {
		return out
	}
	out.TCP.InPerSec = float64(cur.tcpIn-prev.tcpIn) / dt
	out.TCP.OutPerSec = float64(cur.tcpOut-prev.tcpOut) / dt
	out.TCP.RetransPerSec = float64(cur.tcpRetrans-prev.tcpRetrans) / dt
	out.UDP.InPerSec = float64(cur.udpIn-prev.udpIn) / dt
	out.UDP.OutPerSec = float64(cur.udpOut-prev.udpOut) / dt
	out.ICMP.InPerSec = float64(cur.icmpIn-prev.icmpIn) / dt
	out.ICMP.OutPerSec = float64(cur.icmpOut-prev.icmpOut) / dt
	out.ARP.InPerSec = float64(cur.arpIn-prev.arpIn) / dt
	out.ARP.OutPerSec = float64(cur.arpOut-prev.arpOut) / dt
	out.ARP.LookupsPerSec = float64(cur.arpLookups-prev.arpLookups) / dt
	return out
}

func protoMapGet(m map[string]int64, keys ...string) uint64 {
	for _, k := range keys {
		if v, ok := m[k]; ok && v >= 0 {
			return uint64(v)
		}
	}
	return 0
}
