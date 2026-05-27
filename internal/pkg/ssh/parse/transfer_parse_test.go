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

package parse

import (
	"testing"
)

func TestParseTransferGluedPortAnyOrder(t *testing.T) {
	cases := [][]string{
		{"-p22", "-i~/.ssh/key", "./local", "user@host:/remote/f"},
		{"./local", "user@host:/remote/f", "-p", "22", "-i", "~/.ssh/key"},
		{"user@host:/remote/f", "-p22", "./local", "-i~/.ssh/key"},
	}
	for _, argv := range cases {
		spec, err := ParseTransfer(argv, OpGet)
		if err != nil {
			t.Fatalf("%v: %v", argv, err)
		}
		if spec.Conn.Port != 22 || spec.Conn.Identity != "~/.ssh/key" {
			t.Fatalf("%v: conn=%+v", argv, spec.Conn)
		}
		if spec.LocalPath != "./local" || spec.RemotePath != "/remote/f" || spec.Conn.Destination != "user@host" {
			t.Fatalf("%v: spec=%+v", argv, spec)
		}
	}
}

func TestParseTransferThreeArgs(t *testing.T) {
	spec, err := ParseTransfer([]string{
		"deploy@10.0.0.5", "/var/log/app.log", "./app.log", "-F", "~/.ssh/config",
	}, OpGet)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Conn.Destination != "deploy@10.0.0.5" || spec.RemotePath != "/var/log/app.log" || spec.LocalPath != "./app.log" {
		t.Fatalf("spec=%+v", spec)
	}
	if spec.Conn.ConfigFile != "~/.ssh/config" {
		t.Fatalf("config=%q", spec.Conn.ConfigFile)
	}
}

func TestParseTransferSessionTwoPaths(t *testing.T) {
	spec, err := ParseTransfer([]string{
		"/var/log/app.log", "./app.log", "--session", "prod",
	}, OpGet)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Session != "prod" || spec.RemotePath != "/var/log/app.log" || spec.LocalPath != "./app.log" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseTransferRemoteFirst(t *testing.T) {
	spec, err := ParseTransfer([]string{
		"user@deploy:/var/log/app.log", "./app.log",
	}, OpGet)
	if err != nil {
		t.Fatal(err)
	}
	if spec.RemotePath != "/var/log/app.log" || spec.LocalPath != "./app.log" || spec.Conn.Destination != "user@deploy" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseTransferPut(t *testing.T) {
	spec, err := ParseTransfer([]string{
		"-l", "root", "-p22", "./bin/app", "deploy@host:/usr/local/bin/app",
	}, OpPut)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Conn.User != "root" || spec.Conn.Port != 22 {
		t.Fatalf("conn=%+v", spec.Conn)
	}
	if spec.LocalPath != "./bin/app" || spec.RemotePath != "/usr/local/bin/app" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseTransferRequiresRemoteOrSession(t *testing.T) {
	_, err := ParseTransfer([]string{"./a", "./b"}, OpGet)
	if err == nil {
		t.Fatal("expected error without session or user@host:/path")
	}
}
