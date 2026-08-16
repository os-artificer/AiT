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

import "testing"

func TestFillDarwinNetstat(t *testing.T) {
	data := []byte(`tcp:
	100 packets sent
	0 data packet sent after flow control
	200 packets received
	3 data packets (300 bytes) retransmitted
	4 connection requests
	5 connection accepts
	9 connections established
udp:
	11 datagrams received
	12 datagrams output
	2 dropped due to no socket
	1 with bad checksum
icmp:
	20 messages received
	21 messages sent
	1 input errors
	2 output errors
`)
	var raw protoRaw
	fillDarwinNetstat(&raw, data)
	if raw.tcpOut != 100 || raw.tcpIn != 200 || raw.tcpRetrans != 3 {
		t.Fatalf("tcp: %+v", raw)
	}
	if raw.tcpOpens != 9 || raw.tcpEstab != 9 {
		t.Fatalf("tcp opens/estab: %+v", raw)
	}
	if raw.udpIn != 11 || raw.udpOut != 12 || raw.udpNoPort != 2 || raw.udpInErr != 1 {
		t.Fatalf("udp: %+v", raw)
	}
	if raw.icmpIn != 20 || raw.icmpOut != 21 || raw.icmpInErr != 1 || raw.icmpOutErr != 2 {
		t.Fatalf("icmp: %+v", raw)
	}
}

func TestFillDarwinNetstat_Singular(t *testing.T) {
	data := []byte(`tcp:
	50 packet sent
		1 data packet sent after flow control
	60 packet received
	2 data packet (8 byte) retransmitted
	3 connection request
	4 connection accept
	8 connection established (including accepts)
udp:
	11 datagrams received
	12 datagrams output
`)
	var raw protoRaw
	fillDarwinNetstat(&raw, data)
	if raw.tcpOut != 50 || raw.tcpIn != 60 {
		t.Fatalf("tcp in/out: %+v", raw)
	}
	if raw.tcpRetrans != 2 || raw.tcpOpens != 7 || raw.tcpEstab != 8 {
		t.Fatalf("tcp rest: %+v", raw)
	}
}

func TestCountEstablishedTCP(t *testing.T) {
	data := []byte(`tcp4       0      0  127.0.0.1.1    127.0.0.1.2        ESTABLISHED
tcp6       0      0  ::1.1          ::1.2              ESTABLISHED
udp4       0      0  *.*                               *.*
tcp4       0      0  127.0.0.1.3    127.0.0.1.4        CLOSE_WAIT
`)
	if got := countEstablishedTCP(data); got != 2 {
		t.Fatalf("got %d", got)
	}
}
