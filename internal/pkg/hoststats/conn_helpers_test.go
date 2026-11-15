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
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/shirou/gopsutil/v4/net"
)

func netConnectionStatStub(status string) net.ConnectionStat {
	return net.ConnectionStat{Status: status}
}

func TestNormalizeConnQuery_Defaults(t *testing.T) {
	q, err := NormalizeConnQuery(ConnQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if q.Proto != ConnProtoTCP {
		t.Fatalf("proto=%q want tcp", q.Proto)
	}
	if q.Limit != DefaultConnLimit {
		t.Fatalf("limit=%d want %d", q.Limit, DefaultConnLimit)
	}
}

func TestNormalizeConnQuery_AllMapsInternally(t *testing.T) {
	q, err := NormalizeConnQuery(ConnQuery{Proto: "ALL", Limit: 9000})
	if err != nil {
		t.Fatal(err)
	}
	if q.Proto != ConnProtoAll {
		t.Fatalf("proto=%q", q.Proto)
	}
	if gopsutilConnKind(q.Proto) != "inet" {
		t.Fatalf("kind=%q want inet", gopsutilConnKind(q.Proto))
	}
	if q.Limit != MaxConnLimit {
		t.Fatalf("limit=%d want %d", q.Limit, MaxConnLimit)
	}
}

func TestNormalizeConnQuery_BadProto(t *testing.T) {
	_, err := NormalizeConnQuery(ConnQuery{Proto: "unix"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeConnQuery_States(t *testing.T) {
	q, err := NormalizeConnQuery(ConnQuery{States: []string{"listen,established", "NONE"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.States) != 3 {
		t.Fatalf("states=%v", q.States)
	}
	_, err = NormalizeConnQuery(ConnQuery{States: []string{"NOT_A_STATE"}})
	if err == nil {
		t.Fatal("expected invalid state")
	}
}

func TestNormalizeConnQuery_StateAliases(t *testing.T) {
	q, err := NormalizeConnQuery(ConnQuery{States: []string{"SYN_RECEIVED,FIN_WAIT_1,CLOSED,SYN_RCVD"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{ConnStatusSynRecv, ConnStatusFinWait1, ConnStatusClose}
	if len(q.States) != len(want) {
		t.Fatalf("states=%v want %v", q.States, want)
	}
	for i := range want {
		if q.States[i] != want[i] {
			t.Fatalf("states=%v want %v", q.States, want)
		}
	}
	q, err = NormalizeConnQuery(ConnQuery{States: []string{ConnStatusDelete}})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.States) != 1 || q.States[0] != ConnStatusDelete {
		t.Fatalf("DELETE: %v", q.States)
	}
}

func TestNormalizeConnQuery_NegativeLimit(t *testing.T) {
	_, err := NormalizeConnQuery(ConnQuery{Limit: -1})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCanonicalConnStatus_WindowsIngestMatch(t *testing.T) {
	c := connStatFromGopsutil(netConnectionStatStub("SYN_RECEIVED"))
	if c.Status != ConnStatusSynRecv {
		t.Fatalf("status=%q want %q", c.Status, ConnStatusSynRecv)
	}
	q := ConnQuery{States: []string{ConnStatusSynRecv}}
	if !connMatchesQuery(c, q) {
		t.Fatal("expected alias ingest to match canonical filter")
	}
	closeRow := ConnStat{Status: ConnStatusClose}
	deleteQ := ConnQuery{States: []string{ConnStatusDelete}}
	if connMatchesQuery(closeRow, deleteQ) {
		t.Fatal("DELETE must not match CLOSE")
	}
	deleteRow := ConnStat{Status: ConnStatusDelete}
	if !connMatchesQuery(deleteRow, deleteQ) {
		t.Fatal("DELETE should match DELETE")
	}
}

func TestConnMatchesQuery_ListenOnly(t *testing.T) {
	listen := ConnStat{Status: ConnStatusListen, LocalPort: 80}
	estab := ConnStat{Status: ConnStatusEstablished, LocalPort: 80}
	q := ConnQuery{States: []string{ConnStatusListen}}
	if !connMatchesQuery(listen, q) {
		t.Fatal("LISTEN should match")
	}
	if connMatchesQuery(estab, q) {
		t.Fatal("ESTABLISHED must not match LISTEN filter")
	}
}

func TestConnMatchesQuery_PortAndState(t *testing.T) {
	c := ConnStat{Status: ConnStatusEstablished, LocalPort: 22, RemotePort: 443, PID: 9}
	q := ConnQuery{Port: 22, States: []string{ConnStatusEstablished}}
	if !connMatchesQuery(c, q) {
		t.Fatal("expected match on local port")
	}
	q.Port = 443
	if !connMatchesQuery(c, q) {
		t.Fatal("expected match on remote port")
	}
	q.Port = 80
	if connMatchesQuery(c, q) {
		t.Fatal("expected no match")
	}
}

func TestSortConnStats_StableOrder(t *testing.T) {
	conns := []ConnStat{
		{Proto: "tcp", Status: "LISTEN", LocalIP: "0.0.0.0", LocalPort: 80, PID: 2},
		{Proto: "tcp", Status: "ESTABLISHED", LocalIP: "10.0.0.1", LocalPort: 1, PID: 1},
		{Proto: "udp", Status: "NONE", LocalIP: "0.0.0.0", LocalPort: 53, PID: 3},
	}
	sortConnStats(conns)
	if conns[0].Proto != "tcp" || conns[0].Status != "ESTABLISHED" {
		t.Fatalf("first=%+v", conns[0])
	}
	if conns[2].Proto != "udp" {
		t.Fatalf("last=%+v", conns[2])
	}
}

func TestFormatConnTable_HeadersAndIPv6(t *testing.T) {
	age := int64(75)
	created := time.Date(2026, 3, 1, 12, 0, 0, 0, time.Local)
	out := FormatConnTable([]ConnStat{{
		Proto: ConnProtoTCP, Status: ConnStatusEstablished,
		LocalIP: "127.0.0.1", LocalPort: 1,
		RemoteIP: "::1", RemotePort: 443,
		PID: 42, Process: "curl", AgeSeconds: &age, CreatedAt: &created,
	}})
	if !strings.Contains(out, "PROTO") || !strings.Contains(out, "AGE") || !strings.Contains(out, "CREATED") {
		t.Fatalf("missing headers: %s", out)
	}
	if !strings.Contains(out, "[::1]:443") {
		t.Fatalf("want bracketed ipv6: %s", out)
	}
	if !strings.Contains(out, "1m15s") {
		t.Fatalf("want compact age: %s", out)
	}
	if !strings.Contains(out, created.Local().Format(connCreatedFmt)) {
		t.Fatalf("want created: %s", out)
	}
}

func TestFormatConnTable_HeaderAlignsColumns(t *testing.T) {
	age := int64(5)
	created := time.Date(2026, 3, 1, 12, 0, 5, 0, time.Local)
	createdCell := created.Local().Format(connCreatedFmt)
	out := FormatConnTable([]ConnStat{{
		Proto: "tcp", Status: "LISTEN",
		LocalIP: "1.1.1.1", LocalPort: 80,
		PID: 1, Process: "a", AgeSeconds: &age, CreatedAt: &created,
	}})
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("lines=%v", lines)
	}
	headerAge := strings.Index(lines[0], "AGE")
	dataAge := strings.Index(lines[1], "5s")
	headerCreated := strings.Index(lines[0], "CREATED")
	dataCreated := strings.Index(lines[1], createdCell)
	if headerAge < 0 || dataAge < 0 || headerCreated < 0 || dataCreated < 0 {
		t.Fatalf("missing AGE/CREATED cells:\n%s", out)
	}
	if headerAge != dataAge {
		t.Fatalf("AGE misaligned: header=%d data=%d\n%s", headerAge, dataAge, out)
	}
	if headerCreated != dataCreated {
		t.Fatalf("CREATED misaligned: header=%d data=%d\n%s", headerCreated, dataCreated, out)
	}
}

func TestFormatConnEndpoint_LongIPv6Truncates(t *testing.T) {
	got := formatConnEndpoint("2001:db8:85a3:8d3:1319:8a2e:0370:7348", 443)
	if utf8.RuneCountInString(got) > connColAddr {
		t.Fatalf("len=%d got=%q", utf8.RuneCountInString(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("want TruncateName ellipsis: %q", got)
	}
}

func TestFormatConnProcess_RuneSafe(t *testing.T) {
	// 19+ runes so TruncateName(18) must shorten on a rune boundary.
	got := formatConnProcess("中文进程名称超长测试额外字符一二三四五")
	if utf8.RuneCountInString(got) > connColProcess {
		t.Fatalf("len=%d got=%q", utf8.RuneCountInString(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("want ellipsis: %q", got)
	}
	if strings.Contains(got, "\xff") {
		t.Fatalf("invalid utf8: %q", got)
	}
}

func TestFormatConnAge(t *testing.T) {
	cases := []struct {
		sec  int64
		want string
	}{
		{5, "5s"},
		{75, "1m15s"},
		{3600, "1h"},
		{3660, "1h1m"},
		{86400, "1d"},
		{90000, "1d1h"},
	}
	for _, tc := range cases {
		sec := tc.sec
		if got := FormatConnAge(&sec); got != tc.want {
			t.Fatalf("age %d: got %q want %q", tc.sec, got, tc.want)
		}
	}
	if FormatConnAge(nil) != "-" {
		t.Fatal("nil age")
	}
}

func TestWriteConnectionsJSON_Envelope(t *testing.T) {
	var buf bytes.Buffer
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := WriteConnectionsJSON(&buf, at, []ConnStat{{Proto: "tcp", Status: "LISTEN"}}); err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["time"]; !ok {
		t.Fatal("missing time")
	}
	if _, ok := got["connections"]; !ok {
		t.Fatal("missing connections")
	}
	if bytes.Contains(buf.Bytes(), []byte(`"net"`)) {
		t.Fatal("traffic net key must not appear")
	}
}

func TestWriteConnectionsJSON_NullAge(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteConnectionsJSON(&buf, time.Now(), []ConnStat{{Proto: "tcp"}}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"age_seconds": null`) {
		t.Fatalf("want null age_seconds: %s", got)
	}
	if !strings.Contains(got, `"created_at": null`) {
		t.Fatalf("want null created_at: %s", got)
	}
}

func TestFormatConnCreated(t *testing.T) {
	if FormatConnCreated(nil) != "-" {
		t.Fatal("nil created")
	}
	ts := time.Date(2026, 3, 1, 12, 0, 0, 0, time.Local)
	want := ts.Local().Format(connCreatedFmt)
	if got := FormatConnCreated(&ts); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestProcessDisplayName_Fallback(t *testing.T) {
	if got := processDisplayName("nginx", []string{"/usr/sbin/nginx"}); got != "nginx" {
		t.Fatalf("name wins: %q", got)
	}
	if got := processDisplayName("", []string{"/usr/bin/python3", "app.py"}); got != "python3" {
		t.Fatalf("cmdline basename: %q", got)
	}
	if got := processDisplayName("", nil); got != "" {
		t.Fatalf("empty: %q", got)
	}
	if got := processDisplayName("", []string{""}); got != "" {
		t.Fatalf("empty argv0: %q", got)
	}
}
