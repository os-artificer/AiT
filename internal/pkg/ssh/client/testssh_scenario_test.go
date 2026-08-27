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

package client

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/testssh"
)

func TestDial_OneShotEcho(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	srv := testssh.Start(t, testssh.Config{})
	c, err := Dial(Options{
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

	stdout, stderr := captureStd(t, func() {
		if err := c.RunRemoteCommand(context.Background(), "echo mock-oneshot"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(stdout, "mock-oneshot") {
		t.Fatalf("stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestDial_BadPassword(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	srv := testssh.Start(t, testssh.Config{})
	_, err := Dial(Options{
		Host:     srv.Host(),
		Port:     srv.Port(),
		User:     srv.User(),
		Password: "not-the-password",
		UseAgent: false,
	})
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestDial_UnknownCommandExit(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	srv := testssh.Start(t, testssh.Config{})
	c, err := Dial(Options{
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

	err = c.RunRemoteCommand(context.Background(), "nosuchcmd")
	if err == nil {
		t.Fatal("expected remote exit error")
	}
	re, ok := err.(*RemoteExitError)
	if !ok || re.Code != 127 {
		t.Fatalf("err=%v", err)
	}
}

func TestDial_ConnectionRefused(t *testing.T) {
	t.Setenv("AIT_SSH_INSECURE", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_AUTH_SOCK", "")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	_, err = Dial(Options{
		Host:     host,
		Port:     port,
		User:     "dev",
		Password: "x",
		UseAgent: false,
	})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func captureStd(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	rout, wout, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rerr, werr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wout, werr
	fn()
	_ = wout.Close()
	_ = werr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	var outBuf, errBuf bytes.Buffer
	_, _ = io.Copy(&outBuf, rout)
	_, _ = io.Copy(&errBuf, rerr)
	return outBuf.String(), errBuf.String()
}
