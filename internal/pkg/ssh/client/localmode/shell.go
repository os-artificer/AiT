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

// Package localmode runs the !ait local PTY shell that an SSH user can spawn
// without dropping their remote connection. It owns the per-shell PTY lifecycle
// (Shell), the local-shell environment setup (Env), and the shell-specific rc
// file preparation (Prepare for bash/zsh).
package localmode

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

const (
	noticeEOL   = "\r\n"
	copyBufSize = 32 * 1024
)

// Shell runs an interactive shell on the local machine via a PTY.
// Use Start to spawn it and Stop to terminate it; methods are safe to call
// from multiple goroutines.
type Shell struct {
	mu      sync.Mutex
	pty     *os.File
	cmd     *exec.Cmd
	done    chan struct{}
	onExit  func()
	initDir string
	closing atomic.Bool
}

// Start spawns the user's default shell in a new PTY sized to cols/rows.
// onExit, when non-nil, is invoked once after the shell process exits (e.g. to
// re-enable the suspended remote output stream).
// Returns an error when a shell is already running, when init script setup fails,
// or when pty.Start fails.
func (l *Shell) Start(cols, rows int, onExit func()) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.pty != nil {
		return fmt.Errorf("local shell already running")
	}

	l.closing.Store(false)
	shell := Default()
	init, extraArgs, extraEnv, err := Prepare(shell)
	if err != nil {
		return fmt.Errorf("prepare local shell init: %w", err)
	}

	if init != nil {
		l.initDir = init.dir
	}

	cmd := exec.Command(shell, extraArgs...)
	cmd.Env = Env(os.Environ())
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

// Stop terminates the shell process (if any) and removes its temporary init dir.
// Safe to call when the shell has already exited.
func (l *Shell) Stop() {
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

// Write forwards user keystrokes to the local PTY.
// Returns an error when the shell is not running or when the PTY write fails.
func (l *Shell) Write(b []byte) error {
	l.mu.Lock()
	ptmx := l.pty
	l.mu.Unlock()
	if ptmx == nil {
		return fmt.Errorf("local shell not running")
	}
	_, err := ptmx.Write(b)
	return err
}

// Resize updates the PTY window size. No-ops when the shell is not running or
// when cols/rows are non-positive.
func (l *Shell) Resize(cols, rows int) error {
	l.mu.Lock()
	ptmx := l.pty
	l.mu.Unlock()
	if ptmx == nil || cols <= 0 || rows <= 0 {
		return nil
	}
	return pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}

// Active reports whether a PTY is currently attached (set by Start; cleared
// by Stop or by the copyOut goroutine on EOF).
func (l *Shell) Active() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pty != nil
}

// Running reports whether the underlying shell process is still alive
// (sends signal 0; tighter than Active when the PTY is still open but the
// process has been killed externally).
func (l *Shell) Running() bool {
	l.mu.Lock()
	ptmx := l.pty
	cmd := l.cmd
	l.mu.Unlock()
	if ptmx == nil || cmd == nil || cmd.Process == nil {
		return false
	}
	return cmd.Process.Signal(syscall.Signal(0)) == nil
}

// SetPTYForTest assigns the PTY field directly; intended only for unit tests
// that need to exercise Active() without spawning a real shell.
func (l *Shell) SetPTYForTest(f *os.File) {
	l.mu.Lock()
	l.pty = f
	l.mu.Unlock()
}

func (l *Shell) copyOut() {
	defer close(l.done)
	buf := make([]byte, copyBufSize)
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
				_, _ = fmt.Fprintf(os.Stderr, "%sAiT: local shell: %v%s", noticeEOL, err, noticeEOL)
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
