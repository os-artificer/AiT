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

package logsrc_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want logsrc.Level
	}{
		{"err", logsrc.LevelError},
		{"error", logsrc.LevelError},
		{"3", logsrc.LevelError},
		{"fault", logsrc.LevelCritical},
		{"debug", logsrc.LevelDebug},
		{"1", logsrc.LevelAlert}, // syslog numeric
	}
	for _, tc := range cases {
		got, err := logsrc.ParseLevel(tc.in)
		if err != nil {
			t.Fatalf("ParseLevel(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseLevel(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
	if _, err := logsrc.ParseLevel("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTimeExpression(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	got, err := logsrc.ParseTimeExpression("1h", now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(now.Add(-time.Hour)) {
		t.Fatalf("got %v", got)
	}
	got, err = logsrc.ParseTimeExpression("today", now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 0 || got.Day() != 15 {
		t.Fatalf("today=%v", got)
	}
	if _, err := logsrc.ParseTimeExpression("bad", now); err == nil {
		t.Fatal("expected error")
	}
}

func TestFilter_KeywordsAND(t *testing.T) {
	f, err := logsrc.NewFilter(logsrc.Query{Keywords: []string{"error", "disk"}, IgnoreCase: true})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Match(logsrc.Record{Message: "Disk ERROR on /dev/sda"}) {
		t.Fatal("expected match")
	}
	if f.Match(logsrc.Record{Message: "error only"}) {
		t.Fatal("expected no match")
	}
}

func TestFilter_RegexCaseInsensitive(t *testing.T) {
	f, err := logsrc.NewFilter(logsrc.Query{Regex: "fail(ed|ure)", IgnoreCase: true})
	if err != nil {
		t.Fatal(err)
	}
	if !f.Match(logsrc.Record{Message: "FAILURE detected"}) {
		t.Fatal("expected match")
	}
}

func TestRing_LastN(t *testing.T) {
	r := logsrc.NewRing(3)
	for i := 1; i <= 5; i++ {
		r.Add(logsrc.Record{Message: string(rune('0' + i))})
	}
	snap := r.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("len=%d", len(snap))
	}
	if snap[0].Message != "3" || snap[2].Message != "5" {
		t.Fatalf("%v", snap)
	}
}

func TestCollectLastN_Reverse(t *testing.T) {
	recs := []logsrc.Record{
		{Message: "a"}, {Message: "b"}, {Message: "c"}, {Message: "d"},
	}
	// newest-first cursor
	cur := &sliceCur{recs: []logsrc.Record{recs[3], recs[2], recs[1], recs[0]}}
	out, err := logsrc.CollectLastN(cur, nil, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].Message != "c" || out[1].Message != "d" {
		t.Fatalf("%+v", out)
	}
}

// sliceCur is a Cursor over a fixed slice, optionally failing at the end.
type sliceCur struct {
	recs []logsrc.Record
	err  error
	i    int
}

func (c *sliceCur) Next(rec *logsrc.Record) bool {
	if c.i >= len(c.recs) {
		return false
	}
	*rec = c.recs[c.i]
	c.i++
	return true
}
func (c *sliceCur) Err() error   { return c.err }
func (c *sliceCur) Close() error { return nil }

func TestParseJournalJSON(t *testing.T) {
	line := []byte(`{"__REALTIME_TIMESTAMP":"1700000000000000","MESSAGE":"hello","PRIORITY":"3","_HOSTNAME":"host","_SYSTEMD_UNIT":"sshd.service","SYSLOG_IDENTIFIER":"sshd","_PID":"42"}`)
	rec, err := logsrc.ParseJournalJSON(line)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Message != "hello" || rec.Level != logsrc.LevelError || rec.PID != 42 {
		t.Fatalf("%+v", rec)
	}
	if rec.Process != "sshd" {
		t.Fatalf("process=%q", rec.Process)
	}
}

func TestParseLogShowNDJSON_ProcessID(t *testing.T) {
	line := []byte(`{"timestamp":"2026-08-16T16:15:45.000000+0800","eventMessage":"connection accepted","processID":70249,"processImagePath":"/usr/libexec/sshd","messageType":"Default"}`)
	rec, err := logsrc.ParseLogShowNDJSON(line)
	if err != nil {
		t.Fatal(err)
	}
	if rec.PID != 70249 {
		t.Fatalf("pid=%d", rec.PID)
	}
	if rec.Process != "sshd" {
		t.Fatalf("process=%q", rec.Process)
	}
	if rec.Message != "connection accepted" {
		t.Fatalf("msg=%q", rec.Message)
	}
}

func TestParseLogShowSyslog_ProcessPID(t *testing.T) {
	line := "2026-08-16 16:15:45.000000+0800 sshd[70249]: connection accepted"
	rec, ok := logsrc.ParseLogShowSyslog(line)
	if !ok {
		t.Fatal("parse failed")
	}
	if rec.PID != 70249 || rec.Process != "sshd" {
		t.Fatalf("pid=%d process=%q", rec.PID, rec.Process)
	}
	if rec.Message != "connection accepted" {
		t.Fatalf("msg=%q", rec.Message)
	}
}

func TestBuildJournalctlArgs(t *testing.T) {
	q := logsrc.Query{Unit: "sshd", MinLevel: logsrc.LevelError}
	args := logsrc.BuildJournalctlArgs(q, true, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--reverse") || !strings.Contains(joined, "-o json") {
		t.Fatalf("%v", args)
	}
	if strings.Contains(joined, "-n ") {
		t.Fatal("must not pass record cap")
	}
}

func TestFileSource_Builtin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	content := "Jan  2 15:04:05 host sshd[1]: hello world\nJan  2 15:04:06 host sshd[1]: goodbye\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	src := &logsrc.FileSource{Paths: []string{path}, Label: "test"}
	cur, err := src.Open(context.Background(), logsrc.Query{})
	if err != nil {
		t.Fatal(err)
	}
	defer cur.Close()
	var got []string
	var rec logsrc.Record
	for cur.Next(&rec) {
		got = append(got, rec.Message)
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestParseKmsgLine(t *testing.T) {
	boot := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	rec, ok := logsrc.ParseKmsgLine("3,123,2500000,-;kernel panic", boot)
	if !ok {
		t.Fatal("expected ok")
	}
	if rec.Level != logsrc.LevelError || rec.Message != "kernel panic" {
		t.Fatalf("%+v", rec)
	}
	// The kmsg timestamp is microseconds since boot, not a Unix time.
	if want := boot.Add(2500 * time.Millisecond); !rec.Time.Equal(want) {
		t.Fatalf("time=%v want %v", rec.Time, want)
	}
}

func TestParseKmsgLine_UnknownBootLeavesTimeZero(t *testing.T) {
	rec, ok := logsrc.ParseKmsgLine("3,123,2500000,-;kernel panic", time.Time{})
	if !ok {
		t.Fatal("expected ok")
	}
	if !rec.Time.IsZero() {
		t.Fatalf("time=%v want zero so that time filters treat it as unknown", rec.Time)
	}
}

func TestResolver_BuiltinNoLookPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	if err := os.WriteFile(path, []byte("2024-01-02T03:04:05Z hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	looks := 0
	r := &logsrc.Resolver{
		LookPath: func(string) (string, error) {
			looks++
			return "", os.ErrNotExist
		},
		Notice: os.Stderr,
	}
	_, cur, err := r.OpenResolved(context.Background(), logsrc.Query{
		Source:  path,
		Backend: logsrc.BackendBuiltin,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cur.Close()
	if looks != 0 {
		t.Fatalf("LookPath called %d times", looks)
	}
}

func TestParseEventXML_ProcessID(t *testing.T) {
	data := []byte(`<Event>
  <System>
    <Provider Name="Application Error"/>
    <EventID>1000</EventID>
    <Level>2</Level>
    <TimeCreated SystemTime="2026-08-16T09:00:00.0000000Z"/>
    <Computer>winhost</Computer>
    <Execution ProcessID="4242" ThreadID="8"/>
  </System>
  <EventData>
    <Data Name="ProcessName">C:\Windows\System32\svchost.exe</Data>
  </EventData>
  <RenderingInfo><Message>fault</Message></RenderingInfo>
</Event>`)
	rec, err := logsrc.ParseEventXML(data)
	if err != nil {
		t.Fatal(err)
	}
	if rec.PID != 4242 {
		t.Fatalf("pid=%d", rec.PID)
	}
	if rec.Process != "svchost.exe" {
		t.Fatalf("process=%q", rec.Process)
	}
	if rec.Unit != "Application Error" {
		t.Fatalf("unit=%q", rec.Unit)
	}
}
