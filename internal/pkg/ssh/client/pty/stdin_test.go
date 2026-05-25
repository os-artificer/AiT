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

package pty

import (
	"io"
	"testing"
	"time"
)

func TestStdinReader_PauseExclusive(t *testing.T) {
	pr, pw := io.Pipe()
	sr := StartFrom(pr)

	if _, err := pw.Write([]byte{'a'}); err != nil {
		t.Fatal(err)
	}
	waitStdinByte(t, sr.Ch, 'a')

	replIn := sr.BeginREPL()
	defer sr.EndREPL()

	go func() { _, _ = pw.Write([]byte{'b'}) }()
	assertNoStdinByte(t, sr.Ch, 80*time.Millisecond)

	buf := make([]byte, 1)
	if _, err := replIn.Read(buf); err != nil {
		t.Fatal(err)
	}
	if buf[0] != 'b' {
		t.Fatalf("repl byte = %q, want b", buf[0])
	}
}

func TestStdinReader_DrainWhilePaused(t *testing.T) {
	pr, pw := io.Pipe()
	sr := StartFrom(pr)

	if _, err := pw.Write([]byte("ab")); err != nil {
		t.Fatal(err)
	}
	waitStdinByte(t, sr.Ch, 'a')
	waitStdinByte(t, sr.Ch, 'b')

	_ = sr.BeginREPL()
	defer sr.EndREPL()

	assertNoStdinByte(t, sr.Ch, 20*time.Millisecond)
}

func TestStdinReader_ResumeAfterPause(t *testing.T) {
	pr, pw := io.Pipe()
	sr := StartFrom(pr)

	sr.BeginREPL()
	sr.EndREPL()

	if _, err := pw.Write([]byte{'z'}); err != nil {
		t.Fatal(err)
	}
	waitStdinByte(t, sr.Ch, 'z')
}

func TestReplStdinReader_UTF8(t *testing.T) {
	ch := make(chan byte, 8)
	for _, b := range []byte("帮助") {
		ch <- b
	}
	close(ch)

	r := NewReplReader(ch)
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "帮助" {
		t.Fatalf("got %q, want 帮助", got)
	}
}

func waitStdinByte(t *testing.T, ch <-chan StdinByte, want byte) {
	t.Helper()

	select {
	case sb := <-ch:
		if sb.Err != nil {
			t.Fatalf("unexpected err: %v", sb.Err)
		}
		if sb.B != want {
			t.Fatalf("byte = %q, want %q", sb.B, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for stdin byte")
	}
}

func assertNoStdinByte(t *testing.T, ch <-chan StdinByte, wait time.Duration) {
	t.Helper()

	select {
	case sb := <-ch:
		t.Fatalf("unexpected byte %#x err=%v", sb.B, sb.Err)
	case <-time.After(wait):
	}
}
