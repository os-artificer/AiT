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

package zmodem

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestNewReceiveSession_FeedIdle(t *testing.T) {
	var ptyOut bytes.Buffer
	sess, err := NewReceiveSession(&ptyOut, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	n, err := sess.Feed([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("feed n = %d", n)
	}
}

func TestNewSendSession_OpensLocalFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/payload"
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	sess, err := NewSendSession(io.Discard, path, "remote.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	if sess == nil {
		t.Fatal("nil session")
	}
}

func TestSession_AbortWritesCancelSeq(t *testing.T) {
	var ptyOut bytes.Buffer
	sess, err := NewReceiveSession(&ptyOut, t.TempDir()+"/out.bin", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess.Abort()
	if ptyOut.Len() == 0 {
		t.Fatal("expected abort sequence on pty out")
	}
}
