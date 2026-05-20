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

package sshclient

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestHostKeyCallbackAcceptNew(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/known_hosts"
	cb, err := hostKeyCallbackAt(path, "example.com", 22, "accept-new", false)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:22")
	if err := cb("example.com:22", addr, signer.PublicKey()); err != nil {
		t.Fatalf("callback: %v", err)
	}
}

func TestHostKeyCallbackAskAccepts(t *testing.T) {
	oldConfirm := hostKeyConfirm
	oldTTY := isInteractiveTTY
	hostKeyConfirm = func(string) bool { return true }
	isInteractiveTTY = func() bool { return true }
	defer func() {
		hostKeyConfirm = oldConfirm
		isInteractiveTTY = oldTTY
	}()

	dir := t.TempDir()
	path := dir + "/known_hosts"
	cb, err := hostKeyCallbackAt(path, "example.com", 60528, "ask", false)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:60528")
	if err := cb("example.com:60528", addr, signer.PublicKey()); err != nil {
		t.Fatalf("callback: %v", err)
	}
}

func TestHostKeyCallbackAskRejects(t *testing.T) {
	oldConfirm := hostKeyConfirm
	oldTTY := isInteractiveTTY
	hostKeyConfirm = func(string) bool { return false }
	isInteractiveTTY = func() bool { return true }
	defer func() {
		hostKeyConfirm = oldConfirm
		isInteractiveTTY = oldTTY
	}()

	dir := t.TempDir()
	path := dir + "/known_hosts"
	cb, err := hostKeyCallbackAt(path, "example.com", 22, "ask", false)
	if err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	addr, _ := net.ResolveTCPAddr("tcp", "example.com:22")
	err = cb("example.com:22", addr, signer.PublicKey())
	if err == nil || !strings.Contains(err.Error(), "verification failed") {
		t.Fatalf("want verification failed, got %v", err)
	}
}
