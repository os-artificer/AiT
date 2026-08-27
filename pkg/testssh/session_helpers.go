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
	"io"
	"net"
	"strings"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

func acceptLoop(s *Server) {
	defer close(s.done)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go serveConn(conn, s.cfg, s.hostSigner, s.clientSigner)
	}
}

func serveConn(nConn net.Conn, cfg Config, hostSigner, clientSigner ssh.Signer) {
	defer nConn.Close()

	sconn, chans, reqs, err := ssh.NewServerConn(nConn, serverConfig(cfg, hostSigner, clientSigner))
	if err != nil {
		return
	}
	defer sconn.Close()
	go ssh.DiscardRequests(reqs)

	for newCh := range chans {
		if newCh.ChannelType() != "session" {
			_ = newCh.Reject(ssh.UnknownChannelType, "only session is supported")
			continue
		}
		ch, in, err := newCh.Accept()
		if err != nil {
			continue
		}
		go handleSession(ch, in, cfg)
	}
}

func handleSession(ch ssh.Channel, in <-chan *ssh.Request, cfg Config) {
	defer ch.Close()

	for req := range in {
		switch req.Type {
		case "exec":
			handleExec(ch, req, cfg)
			return
		case "subsystem":
			if handleSubsystem(ch, req, cfg) {
				return
			}
		case "pty-req", "env", "window-change":
			if req.WantReply {
				_ = req.Reply(true, nil)
			}
		case "shell":
			if req.WantReply {
				_ = req.Reply(true, nil)
			}
			go drainSessionRequests(in)
			runInteractiveShell(ch, cfg)
			_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
			return
		default:
			if req.WantReply {
				_ = req.Reply(false, nil)
			}
		}
	}
}

func handleExec(ch ssh.Channel, req *ssh.Request, cfg Config) {
	var payload struct {
		Command string
	}
	_ = ssh.Unmarshal(req.Payload, &payload)
	if req.WantReply {
		_ = req.Reply(true, nil)
	}

	stdout, stderr, code := runCommand(cfg.RootDir, cfg.Hostname, payload.Command)
	_, _ = io.WriteString(ch, stdout)
	if stderr != "" {
		_, _ = io.WriteString(ch.Stderr(), stderr)
	}
	_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{uint32(code)}))
}

func handleSubsystem(ch ssh.Channel, req *ssh.Request, cfg Config) bool {
	var payload struct {
		Name string
	}
	_ = ssh.Unmarshal(req.Payload, &payload)
	if !strings.EqualFold(payload.Name, "sftp") {
		if req.WantReply {
			_ = req.Reply(false, nil)
		}
		return false
	}
	if req.WantReply {
		_ = req.Reply(true, nil)
	}
	srv, err := sftp.NewServer(ch, sftp.WithServerWorkingDirectory(cfg.RootDir))
	if err != nil {
		return true
	}
	_ = srv.Serve()
	_ = srv.Close()
	return true
}
