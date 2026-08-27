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

package transfer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/transfer"
	"github.com/os-artificer/AiT/pkg/testssh"
)

func TestGetPut_MockRoundtrip(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	srv := testssh.Start(t, testssh.Config{})
	c, err := client.Dial(client.Options{
		Host:     srv.Host(),
		Port:     srv.Port(),
		User:     srv.User(),
		Password: srv.Password(),
		UseAgent: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	localDir := t.TempDir()
	got := filepath.Join(localDir, "welcome.out")
	if err := transfer.Get(context.Background(), c.Conn(), srv.RemotePath(testssh.WelcomeName), got); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != testssh.WelcomeBody {
		t.Fatalf("get body=%q", body)
	}

	src := filepath.Join(localDir, "up.txt")
	if err := os.WriteFile(src, []byte("put-roundtrip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	remote := srv.RemotePath("from-local.txt")
	if err := transfer.Put(context.Background(), c.Conn(), src, remote); err != nil {
		t.Fatal(err)
	}
	remoteBody, err := os.ReadFile(remote)
	if err != nil {
		t.Fatal(err)
	}
	if string(remoteBody) != "put-roundtrip\n" {
		t.Fatalf("put body=%q", remoteBody)
	}
}

func TestGet_MockMissing(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	srv := testssh.Start(t, testssh.Config{})
	c, err := client.Dial(client.Options{
		Host:     srv.Host(),
		Port:     srv.Port(),
		User:     srv.User(),
		Password: srv.Password(),
		UseAgent: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err = transfer.Get(context.Background(), c.Conn(), srv.RemotePath("missing.bin"), filepath.Join(t.TempDir(), "x"))
	if err == nil {
		t.Fatal("expected missing remote file error")
	}
}
