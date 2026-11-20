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

package keepalive

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDecide_MemoryDoesNotThrottle(t *testing.T) {
	usage := Usage{MemoryBytes: 200}
	limits := Limits{MemoryBytes: 100}
	if got := Decide(usage, limits, HardCaps{}, Hold{Memory: time.Second}, Sustain); got != ActionNone {
		t.Fatalf("early=%v", got)
	}
	if got := Decide(usage, limits, HardCaps{}, Hold{Memory: Sustain}, Sustain); got != ActionKill {
		t.Fatalf("sustain=%v", got)
	}
	if got := Decide(usage, limits, HardCaps{Memory: true}, Hold{Memory: Sustain}, Sustain); got != ActionNone {
		t.Fatalf("capped=%v", got)
	}
}

func TestDecide_IOUtilIgnoresIdleProcess(t *testing.T) {
	usage := Usage{DiskIOPercent: 99, IOKnown: true}
	limits := Limits{DiskIOPercent: 10}
	if got := Decide(usage, limits, HardCaps{}, Hold{IO: Sustain}, Sustain); got != ActionNone {
		t.Fatalf("idle=%v", got)
	}
	usage.ReadBPS = 1
	if got := Decide(usage, limits, HardCaps{}, Hold{IO: Sustain}, Sustain); got != ActionKill {
		t.Fatalf("hot=%v", got)
	}
}

func TestThrottlePause_FitsSample(t *testing.T) {
	pause := throttlePause(Usage{CPUPercent: 80}, Limits{CPUPercent: 10}, HardCaps{})
	if pause <= 0 || pause > SampleEvery {
		t.Fatalf("pause=%s", pause)
	}
	run := SampleEvery - pause
	ratio := float64(run) / float64(SampleEvery)
	if ratio < 0.1 || ratio > 0.2 {
		t.Fatalf("run ratio=%v pause=%s", ratio, pause)
	}
}

func TestWriteCPUMax_AboveOneCore(t *testing.T) {
	dir := t.TempDir()
	if err := WriteCPUMax(dir, 150); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cpu.max"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "150000 100000\n" {
		t.Fatalf("cpu.max=%q", data)
	}
}

func TestWriteIOMax_AndMemoryMax(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMemoryMax(dir, 4096); err != nil {
		t.Fatal(err)
	}
	if err := WriteIOMax(dir, 8, 1, 10, 0); err != nil {
		t.Fatal(err)
	}
	io, err := os.ReadFile(filepath.Join(dir, "io.max"))
	if err != nil {
		t.Fatal(err)
	}
	if string(io) != "8:1 rbps=10 wbps=max\n" {
		t.Fatalf("io.max=%q", io)
	}
}

func TestCPURateFromCores_AboveMachine(t *testing.T) {
	if _, ok := cpuRateFromCores(150, 1); ok {
		t.Fatal("150% of one core on a 1-core machine must not be a Windows hard cap")
	}
	rate, ok := cpuRateFromCores(150, 8)
	if !ok || rate <= 100 {
		t.Fatalf("rate=%d ok=%v", rate, ok)
	}
}

func TestDirUsage_BudgetIsOver(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("ab"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := DirUsage(dir, 1); err != ErrWalkCapped {
		t.Fatalf("err=%v", err)
	}
}

func TestPipeName_DependsOnRoot(t *testing.T) {
	a := pipeName("/tmp/one")
	b := pipeName("/tmp/two")
	if a == b || a != pipeName("/tmp/one") {
		t.Fatalf("a=%s b=%s", a, b)
	}
}

func TestReconcile_LeavesUnchanged(t *testing.T) {
	var mu sync.Mutex
	var starts int
	d := &daemon{
		parent: context.Background(),
		sup: &Supervisor{
			Sleep: func(time.Duration) {},
			Run: func(ctx context.Context, unit Unit) (int, error) {
				mu.Lock()
				starts++
				mu.Unlock()
				<-ctx.Done()
				return 0, ErrOperatorStop
			},
		},
		slots: map[string]*slot{},
	}
	unit := Unit{Name: "demo", Command: []string{"sleep"}}
	d.reconcile([]Unit{unit})
	time.Sleep(20 * time.Millisecond)
	d.reconcile([]Unit{unit})
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	got := starts
	mu.Unlock()
	d.stopAll()
	if got != 1 {
		t.Fatalf("starts=%d", got)
	}
}

func TestWatch_SleepHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sup := &Supervisor{
		Now: time.Now,
		Run: func(ctx context.Context, unit Unit) (int, error) {
			cancel()
			return 1, nil
		},
	}
	start := time.Now()
	if _, err := sup.Watch(ctx, Unit{Name: "demo", Command: []string{"false"}}); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("backoff sleep ignored cancel")
	}
}

func TestServe_BadListen(t *testing.T) {
	dir := t.TempDir()
	err := Serve(context.Background(), dir, "127.0.0.1:99999")
	if err == nil {
		t.Fatal("expected listen error")
	}
}

func TestServe_StatusSample(t *testing.T) {
	prev := sampleHook
	sampleHook = func(int, *ReadWrite) Usage { return Usage{CPUPercent: 12.5} }
	t.Cleanup(func() { sampleHook = prev })

	dir := t.TempDir()
	cfg, _, _ := Layout(dir)
	err := Save(cfg, File{Units: []Unit{{
		Name: "demo", Command: []string{"sleep", "30"},
		Limits: Limits{CPUPercent: 20, DiskIOPercent: 80},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, dir, "127.0.0.1:0") }()
	deadline := time.Now().Add(3 * time.Second)
	var text string
	for time.Now().Before(deadline) {
		text, err = QueryStatus(dir)
		if err == nil && strings.Contains(text, "cpu=12.5%") && strings.Contains(text, "io=unsupported") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return")
	}
	if !strings.Contains(text, "cpu=12.5%") || !strings.Contains(text, "io=unsupported") {
		t.Fatalf("status=%q err=%v", text, err)
	}
}

func TestServe_CancelReturns(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, dir, "127.0.0.1:0") }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not return after cancel")
	}
}

func TestFormatStatusLine_Samples(t *testing.T) {
	line := formatStatusLine("demo", stateRunning, Unit{Limits: Limits{CPUPercent: 20, DiskIOPercent: 80}}, Usage{CPUPercent: 10})
	if !strings.Contains(line, "demo running") || !strings.Contains(line, "cpu=10.0%") || !strings.Contains(line, "io=unsupported") {
		t.Fatalf("line=%q", line)
	}
}

func TestReconcile_StaleGenerationKeepsRunning(t *testing.T) {
	var mu sync.Mutex
	var starts int
	d := &daemon{
		parent: context.Background(),
		sup: &Supervisor{
			Sleep: func(time.Duration) {},
			Run: func(ctx context.Context, unit Unit) (int, error) {
				mu.Lock()
				starts++
				mu.Unlock()
				<-ctx.Done()
				return 0, ErrOperatorStop
			},
		},
		slots: map[string]*slot{},
	}
	unit := Unit{Name: "demo", Command: []string{"sleep"}}
	d.reconcile([]Unit{unit})
	waitStarts(t, &mu, &starts, 1)
	next := unit
	next.Command = []string{"sleep", "2"}
	d.reconcile([]Unit{next})
	waitStarts(t, &mu, &starts, 2)
	time.Sleep(30 * time.Millisecond)
	d.mu.Lock()
	sl := d.slots["demo"]
	running, state, gen := sl.running, sl.state, sl.gen
	d.mu.Unlock()
	d.stopAll()
	if !running || state != stateRunning || gen < 2 {
		t.Fatalf("running=%v state=%s gen=%d", running, state, gen)
	}
}

func waitStarts(t *testing.T, mu *sync.Mutex, starts *int, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := *starts
		mu.Unlock()
		if got >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	mu.Lock()
	got := *starts
	mu.Unlock()
	t.Fatalf("starts=%d want %d", got, want)
}
