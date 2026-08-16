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
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

const netAggregateName = "all"

type netCounters struct {
	name                  string
	rx, tx, errIn, errOut uint64
	dropIn, dropOut       uint64
}

func defaultMount() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

func fillCPUStat(ctx context.Context, snap *Snapshot) error {
	cpuPct, err := cpu.PercentWithContext(ctx, DefaultCPUInterval, false)
	if err != nil {
		return fmt.Errorf("cpu: %w", err)
	}
	if len(cpuPct) > 0 {
		snap.CPU.Percent = cpuPct[0]
	}
	return nil
}

func fillMemStat(ctx context.Context, snap *Snapshot) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("mem: %w", err)
	}
	snap.Mem = MemStat{
		Total:       vm.Total,
		Used:        vm.Used,
		Available:   vm.Available,
		UsedPercent: vm.UsedPercent,
	}
	return nil
}

func fillDiskUsage(ctx context.Context, snap *Snapshot) error {
	mount := defaultMount()
	usage, err := disk.UsageWithContext(ctx, mount)
	if err != nil {
		return fmt.Errorf("disk usage: %w", err)
	}
	snap.Disk.Mount = mount
	snap.Disk.Total = usage.Total
	snap.Disk.Used = usage.Used
	snap.Disk.UsedPercent = usage.UsedPercent
	return nil
}

func sumDiskIO(ctx context.Context) (read, write uint64) {
	diskIO, _ := disk.IOCountersWithContext(ctx)
	for _, io := range diskIO {
		read += io.ReadBytes
		write += io.WriteBytes
	}
	return read, write
}

func sampleNetCounters(ctx context.Context) netCounters {
	out := netCounters{name: netAggregateName}
	netIO, _ := net.IOCountersWithContext(ctx, false)
	if len(netIO) == 0 {
		return out
	}
	n := netIO[0]
	out.rx, out.tx = n.BytesRecv, n.BytesSent
	out.errIn, out.errOut = n.Errin, n.Errout
	out.dropIn, out.dropOut = n.Dropin, n.Dropout
	if n.Name != "" {
		out.name = n.Name
	}
	return out
}

func applyNetCounters(snap *Snapshot, n netCounters) {
	snap.Net.Name = n.name
	snap.Net.ErrIn = n.errIn
	snap.Net.ErrOut = n.errOut
	snap.Net.DropIn = n.dropIn
	snap.Net.DropOut = n.dropOut
}

func processCmdline(ctx context.Context, p *process.Process) string {
	parts, err := p.CmdlineSliceWithContext(ctx)
	if err != nil || len(parts) == 0 {
		return ""
	}
	return TruncateName(strings.Join(parts, " "), ProcCmdlineMaxRunes)
}
