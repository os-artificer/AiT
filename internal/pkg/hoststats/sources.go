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

import "runtime"

// ListSources returns collector capability descriptions for this platform.
func ListSources() []SourceDesc {
	return []SourceDesc{
		{ID: IDCPU, Name: "CPU usage", Status: "available", Notes: "aggregate percent via gopsutil"},
		{ID: IDMem, Name: "Memory usage", Status: "available", Notes: "virtual memory used/total"},
		{ID: IDDisk, Name: "Disk usage and IO", Status: "available", Notes: "default mount + IO byte rates"},
		{ID: IDNet, Name: "Network traffic", Status: "available", Notes: "aggregate iface rates, errs, drops"},
		{ID: IDProto, Name: "Protocol stats", Status: "available", Notes: protoNotes()},
		{
			ID:     IDProc,
			Name:   "Process list",
			Status: "available",
			Notes:  procNotes(),
		},
	}
}

func procNotes() string {
	base := "top processes by CPU/MEM/DISK + PEER listen/remote; default limit 10"
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		return base + "; TCP peer summary best-effort"
	default:
		return base + "; limited support on " + runtime.GOOS
	}
}

func protoNotes() string {
	switch runtime.GOOS {
	case "linux":
		return "TCP/UDP/ICMP via /proc; ARP table + lookup rates (best-effort)"
	case "darwin":
		return "TCP/UDP/ICMP via netstat -s; TCP estab via netstat -an; ARP via arp -an (no lookup rate)"
	case "windows":
		return "TCP/UDP/ICMP via iphlpapi (IPv4+IPv6); ARP via GetIpNetTable (no lookup rate)"
	default:
		return "protocol counters unavailable on " + runtime.GOOS
	}
}
