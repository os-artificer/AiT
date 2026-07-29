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
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Collector samples host metrics and keeps prior counters for rate math.
type Collector struct {
	prevDiskRead  uint64
	prevDiskWrite uint64
	prevNetRx     uint64
	prevNetTx     uint64
	prevProcIO    map[int32]uint64
	prevTime      time.Time
	havePrev      bool
}

// NewCollector returns a Collector ready for Sample calls.
func NewCollector() *Collector {
	return &Collector{
		prevProcIO: make(map[int32]uint64),
	}
}

// Sample gathers a Snapshot. opts controls process list sorting/limit.
// The first call may report zero IO rates until a prior sample exists.
func (c *Collector) Sample(ctx context.Context, opts ProcessQuery) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	now := time.Now()
	snap := Snapshot{Time: now}

	cpuPct, err := cpu.PercentWithContext(ctx, DefaultCPUInterval, false)
	if err != nil {
		return Snapshot{}, fmt.Errorf("cpu: %w", err)
	}
	if len(cpuPct) > 0 {
		snap.CPU.Percent = cpuPct[0]
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("mem: %w", err)
	}
	snap.Mem = MemStat{
		Total:       vm.Total,
		Used:        vm.Used,
		Available:   vm.Available,
		UsedPercent: vm.UsedPercent,
	}

	mount := defaultMount()
	usage, err := disk.UsageWithContext(ctx, mount)
	if err != nil {
		return Snapshot{}, fmt.Errorf("disk usage: %w", err)
	}
	snap.Disk.Mount = mount
	snap.Disk.Total = usage.Total
	snap.Disk.Used = usage.Used
	snap.Disk.UsedPercent = usage.UsedPercent

	diskIO, _ := disk.IOCountersWithContext(ctx)
	var dRead, dWrite uint64
	for _, io := range diskIO {
		dRead += io.ReadBytes
		dWrite += io.WriteBytes
	}

	netIO, _ := net.IOCountersWithContext(ctx, false)
	var nRx, nTx, errIn, errOut, dropIn, dropOut uint64
	netName := "all"
	if len(netIO) > 0 {
		nRx = netIO[0].BytesRecv
		nTx = netIO[0].BytesSent
		errIn = netIO[0].Errin
		errOut = netIO[0].Errout
		dropIn = netIO[0].Dropin
		dropOut = netIO[0].Dropout
		if netIO[0].Name != "" {
			netName = netIO[0].Name
		}
	}
	snap.Net.Name = netName
	snap.Net.ErrIn = errIn
	snap.Net.ErrOut = errOut
	snap.Net.DropIn = dropIn
	snap.Net.DropOut = dropOut

	dt := now.Sub(c.prevTime).Seconds()
	if c.havePrev && dt > 0 {
		snap.Disk.ReadBytesPerSec = float64(dRead-c.prevDiskRead) / dt
		snap.Disk.WriteBytesPerSec = float64(dWrite-c.prevDiskWrite) / dt
		snap.Net.RxBytesPerSec = float64(nRx-c.prevNetRx) / dt
		snap.Net.TxBytesPerSec = float64(nTx-c.prevNetTx) / dt
	}

	procs, err := c.sampleProcesses(ctx, opts, dt)
	if err != nil {
		return Snapshot{}, err
	}
	if procs == nil {
		procs = []ProcStat{}
	}
	snap.Processes = procs

	c.prevDiskRead = dRead
	c.prevDiskWrite = dWrite
	c.prevNetRx = nRx
	c.prevNetTx = nTx
	c.prevTime = now
	c.havePrev = true

	return snap, nil
}

func (c *Collector) sampleProcesses(ctx context.Context, opts ProcessQuery, dt float64) ([]ProcStat, error) {
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		// Soft-fail: restricted environments may deny process enumeration.
		return nil, nil
	}

	nextIO := make(map[int32]uint64, len(pids))
	raw := make([]ProcStat, 0, len(pids))

	for _, pid := range pids {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		p, err := process.NewProcessWithContext(ctx, pid)
		if err != nil {
			continue
		}
		name, _ := p.NameWithContext(ctx)
		cpuPct, _ := p.CPUPercentWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)
		diskRate := c.diskRateForPID(pid, p, ctx, dt, nextIO)
		raw = append(raw, ProcStat{
			PID:             pid,
			Name:            name,
			CPUPercent:      cpuPct,
			MemPercent:      float64(memPct),
			DiskBytesPerSec: diskRate,
			NetBytesPerSec:  0, // no portable per-process net; stay 0
		})
	}
	c.prevProcIO = nextIO
	return SortAndLimitProcesses(raw, opts), nil
}

func (c *Collector) diskRateForPID(pid int32, p *process.Process, ctx context.Context, dt float64, nextIO map[int32]uint64) float64 {
	io, err := p.IOCountersWithContext(ctx)
	if err != nil {
		return 0
	}
	total := io.ReadBytes + io.WriteBytes
	nextIO[pid] = total
	prev, ok := c.prevProcIO[pid]
	if !ok || dt <= 0 {
		return 0
	}
	return float64(total-prev) / dt
}
