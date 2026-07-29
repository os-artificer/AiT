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
	CPU       CPUStat    `json:"cpu"`
	Mem       MemStat    `json:"mem"`
	Disk      DiskStat   `json:"disk"`
	Net       NetStat    `json:"net"`
	Processes []ProcStat `json:"processes,omitempty"`
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

// ProcStat is one process row for lists and the dashboard panel.
type ProcStat struct {
	PID             int32   `json:"pid"`
	Name            string  `json:"name"`
	CPUPercent      float64 `json:"cpu_percent"`
	MemPercent      float64 `json:"mem_percent"`
	DiskBytesPerSec float64 `json:"disk_bytes_per_sec"`
	NetBytesPerSec  float64 `json:"net_bytes_per_sec"`
}

// ProcessQuery controls process list sorting and truncation.
type ProcessQuery struct {
	// Limit is max rows (default DefaultProcessLimit; capped at MaxProcessLimit).
	Limit int
	// Sort is primary key: cpu|mem|net|disk (default cpu). Secondary keys follow
	// the fixed order CPU → MEM → NET → DISK excluding the primary.
	Sort string
}

// SourceDesc describes a metrics collector for `aithost sources`.
type SourceDesc struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}
