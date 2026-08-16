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

import "fmt"

// ProtoSection formats TCP/UDP/ARP/ICMP protocol metrics.
type ProtoSection struct{}

// ID implements Section.
func (ProtoSection) ID() string { return IDProto }

// Human implements Section.
func (ProtoSection) Human(s Snapshot) string {
	p := s.Proto
	return fmt.Sprintf(
		"tcp: ↓%s/s ↑%s/s  estab %d  retrans %s/s  opens %d\n"+
			"udp: ↓%s/s ↑%s/s  noport %d  inerr %d\n"+
			"arp: ↓%s/s ↑%s/s  ents %d  incomp %d  lookup %s/s\n"+
			"icmp: ↓%s/s ↑%s/s  inerr %d  outerr %d",
		FormatPacketRate(p.TCP.InPerSec), FormatPacketRate(p.TCP.OutPerSec),
		p.TCP.Estab, FormatPacketRate(p.TCP.RetransPerSec), p.TCP.Opens,
		FormatPacketRate(p.UDP.InPerSec), FormatPacketRate(p.UDP.OutPerSec),
		p.UDP.NoPort, p.UDP.InErr,
		FormatPacketRate(p.ARP.InPerSec), FormatPacketRate(p.ARP.OutPerSec),
		p.ARP.Entries, p.ARP.Incomplete, FormatPacketRate(p.ARP.LookupsPerSec),
		FormatPacketRate(p.ICMP.InPerSec), FormatPacketRate(p.ICMP.OutPerSec),
		p.ICMP.InErr, p.ICMP.OutErr,
	)
}

// JSONValue implements Section.
func (ProtoSection) JSONValue(s Snapshot) any { return s.Proto }
