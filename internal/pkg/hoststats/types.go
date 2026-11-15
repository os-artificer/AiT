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

// Package hoststats samples host CPU, memory, disk, network, and processes.
package hoststats

import "time"

// Snapshot is a point-in-time host metrics sample.
type Snapshot struct {
	Time      time.Time  `json:"time"`
	Host      HostIdent  `json:"host"`
	CPU       CPUStat    `json:"cpu"`
	Mem       MemStat    `json:"mem"`
	Disk      DiskStat   `json:"disk"`
	Net       NetStat    `json:"net"`
	Proto     ProtoStats `json:"proto"`
	Processes []ProcStat `json:"processes,omitempty"`
}

// HostIdent is local machine and logged-in account identity for headers and JSON.
type HostIdent struct {
	Hostname        string `json:"hostname,omitempty"`
	User            string `json:"user,omitempty"`
	UID             string `json:"uid,omitempty"`
	OS              string `json:"os,omitempty"`
	Arch            string `json:"arch,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	Kernel          string `json:"kernel,omitempty"`
	CPUs            int    `json:"cpus,omitempty"`
}

// CPUStat holds aggregate CPU usage.
type CPUStat struct {
	Percent float64 `json:"percent"`
}

// MemStat holds virtual memory usage.
type MemStat struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskStat holds root (or default) mount usage and IO rates.
type DiskStat struct {
	Mount            string  `json:"mount"`
	Total            uint64  `json:"total"`
	Used             uint64  `json:"used"`
	UsedPercent      float64 `json:"used_percent"`
	ReadBytesPerSec  float64 `json:"read_bytes_per_sec"`
	WriteBytesPerSec float64 `json:"write_bytes_per_sec"`
}

// NetStat holds aggregate interface traffic rates and error/drop counters.
type NetStat struct {
	Name          string  `json:"name"`
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
	ErrIn         uint64  `json:"err_in"`
	ErrOut        uint64  `json:"err_out"`
	DropIn        uint64  `json:"drop_in"`
	DropOut       uint64  `json:"drop_out"`
}

// ProtoStats holds TCP/UDP/ARP/ICMP packet rates and protocol-specific gauges.
type ProtoStats struct {
	TCP  TCPStat  `json:"tcp"`
	UDP  UDPStat  `json:"udp"`
	ARP  ARPStat  `json:"arp"`
	ICMP ICMPStat `json:"icmp"`
}

// TCPStat holds TCP segment rates and connection features.
type TCPStat struct {
	InPerSec      float64 `json:"in_per_sec"`
	OutPerSec     float64 `json:"out_per_sec"`
	Estab         uint64  `json:"estab"`
	RetransPerSec float64 `json:"retrans_per_sec"`
	Opens         uint64  `json:"opens"`
}

// UDPStat holds UDP datagram rates and error features.
type UDPStat struct {
	InPerSec  float64 `json:"in_per_sec"`
	OutPerSec float64 `json:"out_per_sec"`
	NoPort    uint64  `json:"noport"`
	InErr     uint64  `json:"inerr"`
}

// ARPStat holds ARP/neighbor rates and table features.
type ARPStat struct {
	InPerSec      float64 `json:"in_per_sec"`
	OutPerSec     float64 `json:"out_per_sec"`
	Entries       uint64  `json:"ents"`
	Incomplete    uint64  `json:"incomp"`
	LookupsPerSec float64 `json:"lookup_per_sec,omitempty"`
}

// ICMPStat holds ICMP message rates and error features.
type ICMPStat struct {
	InPerSec  float64 `json:"in_per_sec"`
	OutPerSec float64 `json:"out_per_sec"`
	InErr     uint64  `json:"inerr"`
	OutErr    uint64  `json:"outerr"`
}

// ProcListen is one listening TCP port and its established connection count.
type ProcListen struct {
	Port  uint32 `json:"port"`
	Conns int    `json:"conns"`
}

// ProcLog is one system-log record attributed to a process.
type ProcLog struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// ProcStat is one process row for lists and the dashboard panel.
type ProcStat struct {
	PID             int32        `json:"pid"`
	PPID            int32        `json:"ppid,omitempty"`
	Name            string       `json:"name"`
	Cmdline         string       `json:"cmdline,omitempty"`
	CPUPercent      float64      `json:"cpu_percent"`
	MemPercent      float64      `json:"mem_percent"`
	DiskBytesPerSec float64      `json:"disk_bytes_per_sec"`
	NetBytesPerSec  float64      `json:"net_bytes_per_sec"`
	Listen          []ProcListen `json:"listen,omitempty"`
	Remotes         []string     `json:"remotes,omitempty"` // "host:port" for client connections
	// LogLine is the newest matching system-log message for the table LOG column.
	LogLine string `json:"log_line,omitempty"`
	// Logs are the newest matching system-log records (oldest first) for detail.
	Logs []ProcLog `json:"logs,omitempty"`
}

// ConnStat is one local network connection row for `aithost net --conn`.
type ConnStat struct {
	Proto      string `json:"proto"`
	Family     string `json:"family"`
	Status     string `json:"status"`
	LocalIP    string `json:"local_ip"`
	LocalPort  uint32 `json:"local_port"`
	RemoteIP   string `json:"remote_ip"`
	RemotePort uint32 `json:"remote_port"`
	PID        int32  `json:"pid"`
	Fd         uint32 `json:"fd"`
	Process    string `json:"process"`
	// CreatedAt is absolute create/start time on Linux when known (fd Ctim, or derived from conntrack age).
	// Paired with AgeSeconds; both null when unknown. No omitempty so JSON emits null.
	CreatedAt *time.Time `json:"created_at"`
	// AgeSeconds is relative lifetime in seconds when known; paired with CreatedAt.
	AgeSeconds *int64 `json:"age_seconds"`
}

// ConnQuery filters and truncates ListConnections results.
type ConnQuery struct {
	// Proto is tcp (default), udp, or all (maps to gopsutil "inet").
	Proto string
	// States, when non-empty, keeps rows whose Status matches (case-insensitive).
	States []string
	// Port, when >0, matches local or remote port.
	Port int
	// PID, when >0, scopes the OS query and filters to that process.
	PID int32
	// Limit caps rows after filter/sort (default DefaultConnLimit; max MaxConnLimit).
	Limit int
}

// ProcessQuery controls process list sorting and truncation.
type ProcessQuery struct {
	// Limit is max rows (default DefaultProcessLimit; capped at MaxProcessLimit).
	// Ignored when All is true.
	Limit int
	// Sort is primary key: cpu|mem|net|disk (default cpu). Secondary keys follow
	// the fixed order CPU → MEM → NET → DISK excluding the primary.
	Sort string
	// KeepPIDs, when non-empty, are preferred in the result (fresh stats) before
	// filling remaining slots from the normal sort order. Used by aihost top
	// search/detail so frozen rows keep refreshing outside the current top-N.
	KeepPIDs []int32
	// All, when true, returns every sampled process (sorted); Limit is ignored.
	// aihost top defaults to All so PROC search covers the full process table.
	All bool
}

// SourceDesc describes a metrics collector for `aihost sources`.
type SourceDesc struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}
