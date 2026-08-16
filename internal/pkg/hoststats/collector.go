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
	"sync"
	"sync/atomic"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"

	"github.com/shirou/gopsutil/v4/process"
)

// Collector samples host metrics and keeps prior counters for rate math.
type Collector struct {
	prevDiskRead  uint64
	prevDiskWrite uint64
	prevNetRx     uint64
	prevNetTx     uint64
	prevProto     protoRaw
	prevProcIO    map[int32]uint64
	prevTime      time.Time
	havePrev      bool

	logMu   sync.Mutex
	logRecs []logsrc.Record
	logAt   time.Time
	logBusy atomic.Bool
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
	snap := Snapshot{Time: now, Host: CollectHostIdent()}
	if err := fillCPUStat(ctx, &snap); err != nil {
		return Snapshot{}, err
	}
	if err := fillMemStat(ctx, &snap); err != nil {
		return Snapshot{}, err
	}
	if err := fillDiskUsage(ctx, &snap); err != nil {
		return Snapshot{}, err
	}

	dRead, dWrite := sumDiskIO(ctx)
	net := sampleNetCounters(ctx)
	applyNetCounters(&snap, net)
	protoCur, _ := sampleProtoRaw(ctx)

	dt := now.Sub(c.prevTime).Seconds()
	if c.havePrev && dt > 0 {
		snap.Disk.ReadBytesPerSec = float64(dRead-c.prevDiskRead) / dt
		snap.Disk.WriteBytesPerSec = float64(dWrite-c.prevDiskWrite) / dt
		snap.Net.RxBytesPerSec = float64(net.rx-c.prevNetRx) / dt
		snap.Net.TxBytesPerSec = float64(net.tx-c.prevNetTx) / dt
	}
	snap.Proto = protoStatsFromRaw(protoCur, c.prevProto, dt, c.havePrev)

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
	c.prevNetRx = net.rx
	c.prevNetTx = net.tx
	c.prevProto = protoCur
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
		ppid, _ := p.PpidWithContext(ctx)
		diskRate := c.diskRateForPID(pid, p, ctx, dt, nextIO)
		raw = append(raw, ProcStat{
			PID:             pid,
			PPID:            ppid,
			Name:            name,
			Cmdline:         processCmdline(ctx, p),
			CPUPercent:      cpuPct,
			MemPercent:      float64(memPct),
			DiskBytesPerSec: diskRate,
			NetBytesPerSec:  0, // no portable per-process net; stay 0
		})
	}
	c.prevProcIO = nextIO
	attachProcPeers(ctx, raw)
	c.attachCachedProcLogs(raw)
	return SortAndLimitProcesses(raw, opts), nil
}

func (c *Collector) attachCachedProcLogs(procs []ProcStat) {
	c.logMu.Lock()
	recs := append([]logsrc.Record(nil), c.logRecs...)
	c.logMu.Unlock()
	attachProcLogs(procs, recs)
	c.refreshProcLogsAsync()
}

func (c *Collector) refreshProcLogsAsync() {
	c.logMu.Lock()
	stale := c.logAt.IsZero() || time.Since(c.logAt) >= procLogMinInterval
	c.logMu.Unlock()
	if !stale {
		return
	}
	if !c.logBusy.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer c.logBusy.Store(false)
		ctx, cancel := context.WithTimeout(context.Background(), procLogCollectTimeout)
		defer cancel()
		recs, err := collectSystemLogs(ctx)
		c.logMu.Lock()
		if err == nil {
			c.logRecs = recs
		}
		c.logAt = time.Now()
		c.logMu.Unlock()
	}()
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
