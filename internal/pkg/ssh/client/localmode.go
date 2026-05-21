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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/creack/pty"
)

const localPromptEOL = "\r\n"

// localShell runs an interactive shell on the local machine via a PTY.
type localShell struct {
	mu      sync.Mutex
	pty     *os.File
	cmd     *exec.Cmd
	done    chan struct{}
	onExit  func()
	initDir string
	closing atomic.Bool
}

func (l *localShell) start(cols, rows int, onExit func()) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.pty != nil {
		return fmt.Errorf("local shell already running")
	}

	l.closing.Store(false)
	shell := defaultLocalShell()
	init, extraArgs, extraEnv, err := prepareLocalShellInit(shell)
	if err != nil {
		return fmt.Errorf("prepare local shell init: %w", err)
	}

	if init != nil {
		l.initDir = init.dir
	}

	cmd := exec.Command(shell, extraArgs...)
	cmd.Env = localShellEnv(os.Environ())
	cmd.Env = append(cmd.Env, extraEnv...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		if init != nil {
			init.remove()
			l.initDir = ""
		}
		return err
	}

	if cols > 0 && rows > 0 {
		_ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
	}

	l.pty = ptmx
	l.cmd = cmd
	l.onExit = onExit
	l.done = make(chan struct{})

	go l.copyOut()
	return nil
}

func (l *localShell) copyOut() {
	defer close(l.done)
	buf := make([]byte, 32*1024)
	for {
		l.mu.Lock()
		ptmx := l.pty
		l.mu.Unlock()
		if ptmx == nil {
			return
		}
		n, err := ptmx.Read(buf)
		if n > 0 {
			_, _ = os.Stdout.Write(buf[:n])
		}
		if err != nil {
			if !l.closing.Load() && !errors.Is(err, io.EOF) {
				_, _ = fmt.Fprintf(os.Stderr, "%sAiT: local shell: %v%s", localPromptEOL, err, localPromptEOL)
			}
			l.mu.Lock()

			onExit := l.onExit
			l.pty = nil
			l.cmd = nil
			l.onExit = nil
			l.mu.Unlock()

			if onExit != nil {
				onExit()
			}
			return
		}
	}
}

func (l *localShell) stop() {
	l.closing.Store(true)

	l.mu.Lock()
	ptmx := l.pty
	cmd := l.cmd
	done := l.done
	initDir := l.initDir
	l.pty = nil
	l.onExit = nil
	l.initDir = ""
	l.mu.Unlock()

	if initDir != "" {
		_ = os.RemoveAll(initDir)
	}

	if ptmx == nil {
		return
	}

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}

	_ = ptmx.Close()
	if done != nil {
		<-done
	}

	l.closing.Store(false)
}

func (l *localShell) write(b []byte) error {
	l.mu.Lock()
	ptmx := l.pty
	l.mu.Unlock()
	if ptmx == nil {
		return fmt.Errorf("local shell not running")
	}
	_, err := ptmx.Write(b)
	return err
}

func (l *localShell) resize(cols, rows int) error {
	l.mu.Lock()
	ptmx := l.pty
	l.mu.Unlock()
	if ptmx == nil || cols <= 0 || rows <= 0 {
		return nil
	}
	return pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}

func (l *localShell) active() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pty != nil
}

// running reports whether the local shell process is still alive.
func (l *localShell) running() bool {
	l.mu.Lock()
	ptmx := l.pty
	cmd := l.cmd
	l.mu.Unlock()
	if ptmx == nil || cmd == nil || cmd.Process == nil {
		return false
	}
	return cmd.Process.Signal(syscall.Signal(0)) == nil
}
