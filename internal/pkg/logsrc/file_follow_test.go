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
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

// Follow must reopen the live path when logrotate replaces the inode.
func TestFileFollow_ReopensAfterRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("line-one\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	src := &logsrc.FileSource{Paths: []string{path}, Label: "test"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cur, err := src.Open(ctx, logsrc.Query{Follow: true})
	if err != nil {
		t.Fatal(err)
	}
	defer cur.Close()

	var rec logsrc.Record
	if !cur.Next(&rec) {
		t.Fatalf("first: %v", cur.Err())
	}
	if rec.Message != "line-one" {
		t.Fatalf("%q", rec.Message)
	}

	rotated := filepath.Join(dir, "app.log.1")
	if err := os.Rename(path, rotated); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("line-two\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for rotated line")
		default:
		}
		if cur.Next(&rec) {
			if rec.Message == "line-two" {
				return
			}
			continue
		}
		if err := cur.Err(); err != nil && err != context.Canceled {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
