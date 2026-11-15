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
	"sort"

	"github.com/shirou/gopsutil/v4/net"
)

// procConn is a minimal TCP connection view for peer aggregation tests.
type procConn struct {
	Pid    int32
	Status string
	LPort  uint32
	RIP    string
	RPort  uint32
}

type procPeerAgg struct {
	Listen  []ProcListen
	Remotes []string
}

type pidBag struct {
	listenPorts map[uint32]struct{}
	estabOnPort map[uint32]int
	remotes     map[string]struct{}
}

func attachProcPeers(ctx context.Context, procs []ProcStat) {
	if len(procs) == 0 {
		return
	}
	stats, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return
	}
	conns := make([]procConn, 0, len(stats))
	for _, s := range stats {
		conns = append(conns, procConn{
			Pid:    s.Pid,
			Status: s.Status,
			LPort:  s.Laddr.Port,
			RIP:    s.Raddr.IP,
			RPort:  s.Raddr.Port,
		})
	}
	peers := aggregateProcPeers(conns)
	for i := range procs {
		if p, ok := peers[procs[i].PID]; ok {
			procs[i].Listen = p.Listen
			procs[i].Remotes = p.Remotes
		}
	}
}

func aggregateProcPeers(conns []procConn) map[int32]procPeerAgg {
	byPID := map[int32]*pidBag{}
	for _, c := range conns {
		if c.Pid <= 0 {
			continue
		}
		b := byPID[c.Pid]
		if b == nil {
			b = &pidBag{
				listenPorts: map[uint32]struct{}{},
				estabOnPort: map[uint32]int{},
				remotes:     map[string]struct{}{},
			}
			byPID[c.Pid] = b
		}
		switch c.Status {
		case ConnStatusListen:
			b.listenPorts[c.LPort] = struct{}{}
		case ConnStatusEstablished:
			recordEstablished(b, c)
		}
	}

	out := make(map[int32]procPeerAgg, len(byPID))
	for pid, b := range byPID {
		if agg, ok := peerAggFromBag(b); ok {
			out[pid] = agg
		}
	}
	return out
}

func recordEstablished(b *pidBag, c procConn) {
	b.estabOnPort[c.LPort]++
	if c.RIP == "" || c.RPort == 0 {
		return
	}
	b.remotes[fmt.Sprintf("%s:%d", c.RIP, c.RPort)] = struct{}{}
}

func peerAggFromBag(b *pidBag) (procPeerAgg, bool) {
	if len(b.listenPorts) > 0 {
		return procPeerAgg{Listen: listenFromBag(b)}, true
	}
	if len(b.remotes) == 0 {
		return procPeerAgg{}, false
	}
	return procPeerAgg{Remotes: sortedKeys(b.remotes)}, true
}

func listenFromBag(b *pidBag) []ProcListen {
	ports := make([]uint32, 0, len(b.listenPorts))
	for p := range b.listenPorts {
		ports = append(ports, p)
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
	listen := make([]ProcListen, 0, len(ports))
	for _, p := range ports {
		listen = append(listen, ProcListen{Port: p, Conns: b.estabOnPort[p]})
	}
	return listen
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
