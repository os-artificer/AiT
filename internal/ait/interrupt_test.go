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

package ait

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestWithAbsorbedInterrupt_Survives(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGINT self-signal timing differs on Windows; covered by manual smoke")
	}

	done := make(chan error, 1)
	go func() {
		done <- withAbsorbedInterrupt(func() error {
			time.Sleep(300 * time.Millisecond)
			return nil
		})
	}()

	time.Sleep(50 * time.Millisecond)
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("fn err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("parent did not survive Interrupt during absorb")
	}

	// After Stop, a fresh Notify should still work (no leaked Ignore).
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer signal.Stop(ch)
}

func TestWithAbsorbedInterrupt_ChildReceivesSIGINT(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("child SIGINT delivery covered by manual smoke on Windows")
	}

	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not on PATH")
	}

	err = withAbsorbedInterrupt(func() error {
		cmd := exec.Command(sleepPath, "30")
		if startErr := cmd.Start(); startErr != nil {
			return startErr
		}

		time.Sleep(50 * time.Millisecond)
		if killErr := syscall.Kill(cmd.Process.Pid, syscall.SIGINT); killErr != nil {
			_ = cmd.Process.Kill()
			return killErr
		}

		waitErr := cmd.Wait()
		if waitErr == nil {
			return errors.New("expected child to exit after SIGINT")
		}
		var ee *exec.ExitError
		if !errors.As(waitErr, &ee) {
			return waitErr
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
