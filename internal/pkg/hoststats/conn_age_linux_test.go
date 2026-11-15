//go:build linux

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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseConntrackLine_DeltaTime(t *testing.T) {
	line := "ipv4     2 tcp      6 431999 ESTABLISHED src=10.0.0.1 dst=1.2.3.4 sport=54321 dport=443 " +
		"src=1.2.3.4 dst=10.0.0.1 sport=443 dport=54321 [ASSURED] mark=0 delta-time=125 use=1"
	tup, age, ok := parseConntrackLine(line)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if age != 125 {
		t.Fatalf("age=%d", age)
	}
	if tup.proto != "tcp" || tup.lip != "10.0.0.1" || tup.rip != "1.2.3.4" || tup.lport != 54321 || tup.rport != 443 {
		t.Fatalf("tuple=%+v", tup)
	}
}

func TestParseConntrackLine_NoDelta(t *testing.T) {
	line := "ipv4     2 tcp      6 431999 ESTABLISHED src=10.0.0.1 dst=1.2.3.4 sport=1 dport=2"
	if _, _, ok := parseConntrackLine(line); ok {
		t.Fatal("expected no parse without delta-time")
	}
}

func TestCreatedAtFromFD_PidZeroSkipped(t *testing.T) {
	if got := createdAtFromFD(0, 3); got != nil {
		t.Fatalf("want nil for pid 0, got %v", got)
	}
}

func TestCreatedAtFromFD_SelfFD(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x")
	if err := os.WriteFile(p, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	pid := int32(os.Getpid())
	fd := uint32(f.Fd())
	created := createdAtFromFD(pid, fd)
	if created == nil {
		t.Fatal("expected created time from self fd")
	}
	if created.IsZero() {
		t.Fatal("created is zero")
	}
}

func TestEnrichConnAge_FDAndConntrackDerived(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x")
	if err := os.WriteFile(p, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// now must be after fd ctime so applyFDConnAge accepts the instant.
	now := time.Now().Add(2 * time.Second)

	fdRow := ConnStat{
		Proto: ConnProtoTCP, PID: int32(os.Getpid()), Fd: uint32(f.Fd()),
		LocalIP: "127.0.0.1", LocalPort: 9,
	}
	trackRow := ConnStat{
		Proto: ConnProtoTCP, PID: 0, Fd: 0,
		LocalIP: "10.0.0.1", LocalPort: 1, RemoteIP: "1.2.3.4", RemotePort: 443,
	}
	track := map[connTuple]int64{
		{proto: "tcp", lip: "10.0.0.1", rip: "1.2.3.4", lport: 1, rport: 443}: 90,
	}

	conns := []ConnStat{fdRow, trackRow}
	enrichConnAgeWith(conns, now, track)

	if conns[0].CreatedAt == nil || conns[0].AgeSeconds == nil {
		t.Fatalf("fd row: created=%v age=%v", conns[0].CreatedAt, conns[0].AgeSeconds)
	}
	if conns[1].CreatedAt == nil || conns[1].AgeSeconds == nil || *conns[1].AgeSeconds != 90 {
		t.Fatalf("track row: created=%v age=%v", conns[1].CreatedAt, conns[1].AgeSeconds)
	}
	wantCreated := now.Add(-90 * time.Second)
	if !conns[1].CreatedAt.Equal(wantCreated) {
		t.Fatalf("derived created=%v want %v", conns[1].CreatedAt, wantCreated)
	}
}

func TestAgeFromConntrack_MatchAndSwap(t *testing.T) {
	track := map[connTuple]int64{
		{proto: "tcp", lip: "10.0.0.1", rip: "1.2.3.4", lport: 1, rport: 443}: 9,
	}
	c := ConnStat{Proto: "tcp", LocalIP: "1.2.3.4", RemoteIP: "10.0.0.1", LocalPort: 443, RemotePort: 1}
	age := ageFromConntrack(track, c)
	if age == nil || *age != 9 {
		t.Fatalf("got %v", age)
	}
}
