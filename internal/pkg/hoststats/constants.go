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

// Metric section IDs (aligned with aithost subcommand names).
const (
	IDCPU  = "cpu"
	IDMem  = "mem"
	IDDisk = "disk"
	IDNet  = "net"
	IDProc = "proc"
)

// Sort key names accepted by ProcessQuery / CLI --sort.
const (
	SortCPU  = IDCPU
	SortMem  = IDMem
	SortNet  = IDNet
	SortDisk = IDDisk
)

// ProcTableHeader is the human-readable process table column header (no trailing newline).
const ProcTableHeader = "PID      NAME                 CPU%   MEM%   DISK/s    NET/s"

// ProcNameColWidth is the display width for process names in tables.
const ProcNameColWidth = 20
