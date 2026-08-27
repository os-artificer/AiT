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

package testssh

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

func TestSSH_ExecEcho(t *testing.T) {
	srv := Start(t, Config{})
	out, code := mustExec(t, srv, srv.Password(), "echo hello-mock")
	if code != 0 || !strings.Contains(out, "hello-mock") {
		t.Fatalf("out=%q code=%d", out, code)
	}
}

func TestSSH_BadPassword(t *testing.T) {
	srv := Start(t, Config{})
	_, err := dialSSH(t, srv, "wrong-password")
	if err == nil {
		t.Fatal("expected auth failure")
	}
}

func TestSSH_RejectPassword(t *testing.T) {
	srv := Start(t, Config{RejectPassword: true})
	_, err := dialSSH(t, srv, DefaultPassword)
	if err == nil {
		t.Fatal("expected reject-password auth failure")
	}
}

func TestListen_RootDirRequired(t *testing.T) {
	_, err := Listen(Config{})
	if err == nil {
		t.Fatal("expected RootDir required")
	}
}

func TestSSH_UnknownCommand(t *testing.T) {
	srv := Start(t, Config{})
	_, code := mustExec(t, srv, srv.Password(), "nosuchcmd")
	if code != 127 {
		t.Fatalf("code=%d want 127", code)
	}
}

func TestSSH_FalseExit(t *testing.T) {
	srv := Start(t, Config{})
	_, code := mustExec(t, srv, srv.Password(), "false")
	if code != 1 {
		t.Fatalf("code=%d want 1", code)
	}
}

func TestSSH_CatWelcome(t *testing.T) {
	srv := Start(t, Config{})
	out, code := mustExec(t, srv, srv.Password(), "cat /welcome.txt")
	if code != 0 || out != WelcomeBody {
		t.Fatalf("out=%q code=%d", out, code)
	}
}

func TestSSH_CatMissing(t *testing.T) {
	srv := Start(t, Config{})
	_, code := mustExec(t, srv, srv.Password(), "cat /nope.txt")
	if code != 1 {
		t.Fatalf("code=%d want 1", code)
	}
}

func TestSSH_SFTPGet(t *testing.T) {
	srv := Start(t, Config{})
	conn := mustDial(t, srv, srv.Password())
	defer conn.Close()

	sf, err := sftp.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	f, err := sf.Open(srv.RemotePath(WelcomeName))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != WelcomeBody {
		t.Fatalf("body=%q", body)
	}
}

func TestSSH_SFTPGetMissing(t *testing.T) {
	srv := Start(t, Config{})
	conn := mustDial(t, srv, srv.Password())
	defer conn.Close()

	sf, err := sftp.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	_, err = sf.Open(srv.RemotePath("missing.bin"))
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestSSH_SFTPPutGet(t *testing.T) {
	srv := Start(t, Config{})
	conn := mustDial(t, srv, srv.Password())
	defer conn.Close()

	sf, err := sftp.NewClient(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	remote := srv.RemotePath("upload.bin")
	w, err := sf.Create(remote)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, "uploaded-bytes\n"); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(remote)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "uploaded-bytes\n" {
		t.Fatalf("disk=%q", got)
	}
}

func TestSSH_PublicKeyAuth(t *testing.T) {
	srv := Start(t, Config{})
	if srv.IdentityPath() == "" {
		t.Fatal("expected generated identity")
	}
	key, err := os.ReadFile(srv.IdentityPath())
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ssh.ClientConfig{
		User:            srv.User(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // loopback mock
	}
	conn, err := ssh.Dial("tcp", srv.Addr(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
}

func mustDial(t *testing.T, srv *Server, password string) *ssh.Client {
	t.Helper()
	conn, err := dialSSH(t, srv, password)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func dialSSH(t *testing.T, srv *Server, password string) (*ssh.Client, error) {
	t.Helper()
	cfg := &ssh.ClientConfig{
		User:            srv.User(),
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // loopback mock
	}
	return ssh.Dial("tcp", srv.Addr(), cfg)
}

func mustExec(t *testing.T, srv *Server, password, cmd string) (string, int) {
	t.Helper()
	conn := mustDial(t, srv, password)
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	var stdout bytes.Buffer
	sess.Stdout = &stdout
	err = sess.Run(cmd)
	code := 0
	if err != nil {
		if ee, ok := err.(*ssh.ExitError); ok {
			code = ee.ExitStatus()
		} else {
			t.Fatal(err)
		}
	}
	return stdout.String(), code
}

func TestSSH_InteractiveEcho(t *testing.T) {
	srv := Start(t, Config{})
	conn := mustDial(t, srv, srv.Password())
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	if err := sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Shell(); err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(stdin, "export PS1='ignored'\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(stdin, "echo hello-interactive\n"); err != nil {
		t.Fatal(err)
	}
	out := readUntilContains(t, stdout, "hello-interactive", 5*time.Second)
	if !strings.Contains(out, "AiT ") {
		t.Fatalf("missing prompt after echo: %q", out)
	}
	if !strings.Contains(out, srv.User()+"@") {
		t.Fatalf("prompt=%q", out)
	}
	if strings.Contains(out, "command not found") {
		t.Fatalf("setup leaked: %q", out)
	}
}

func readUntilContains(t *testing.T, r io.Reader, substr string, timeout time.Duration) string {
	t.Helper()

	done := make(chan string, 1)
	go func() {
		var b strings.Builder
		buf := make([]byte, 256)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				b.Write(buf[:n])
				if strings.Contains(b.String(), substr) {
					done <- b.String()
					return
				}
			}
			if err != nil {
				done <- b.String()
				return
			}
		}
	}()

	select {
	case s := <-done:
		if !strings.Contains(s, substr) {
			t.Fatalf("got %q want substring %q", s, substr)
		}
		return s
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for %q", substr)
	}
	return ""
}

func TestRemotePath_Join(t *testing.T) {
	dir := t.TempDir()
	srv := &Server{cfg: Config{RootDir: dir}}
	got := srv.RemotePath("a.txt")
	if got != filepath.Join(dir, "a.txt") {
		t.Fatalf("got %q", got)
	}
}
