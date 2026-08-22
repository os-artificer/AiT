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

import "time"

const (
	// DefaultProcessLimit is the default number of processes returned.
	DefaultProcessLimit = 10
	// MaxProcessLimit caps --limit to avoid unbounded output.
	MaxProcessLimit = 500
	// DefaultCPUInterval is the wait used for CPU percent sampling (0 = non-blocking / cached).
	DefaultCPUInterval time.Duration = 0
)

// Metric section IDs (aligned with aihost subcommand names).
const (
	IDCPU   = "cpu"
	IDMem   = "mem"
	IDDisk  = "disk"
	IDNet   = "net"
	IDProto = "proto"
	IDProc  = "proc"
)

// Sort key names accepted by ProcessQuery / CLI --sort.
const (
	SortCPU  = IDCPU
	SortMem  = IDMem
	SortNet  = IDNet
	SortDisk = IDDisk
)

// ProcNameColWidth is the display width for process names in tables.
const ProcNameColWidth = 20

// ProcPeerColWidth is the display width for the PEER column.
const ProcPeerColWidth = 24

// ProcLogColWidth is the default display width for the LOG column.
const ProcLogColWidth = 36

// ProcLogTimeLayout is the HH:MM:SS stamp used in the PROC detail log list.
const ProcLogTimeLayout = "15:04:05"

const (
	procPIDColWidth    = 8
	procMetricColWidth = 5
	procDiskColWidth   = 8
	procColSep         = 1
	procMetricSep      = 2
	procLogSep         = 2
)

// ProcRowMetaWidth is the fixed prefix of a process row before the LOG column
// (PID/NAME/CPU/MEM/DISK/PEER plus separators).
const ProcRowMetaWidth = procPIDColWidth + procColSep +
	ProcNameColWidth + procColSep +
	procMetricColWidth + procMetricSep +
	procMetricColWidth + procMetricSep +
	procDiskColWidth + procMetricSep +
	ProcPeerColWidth + procLogSep

// ProcLogDetailMax is how many matching log records are kept per process for detail.
const ProcLogDetailMax = 12

// ProcCmdlineMaxRunes caps stored cmdline length for match/detail.
const ProcCmdlineMaxRunes = 120
