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

package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aissh/shell/builtin"
	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestExecuteLine_LoginNotFound(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "login" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("PATH login must not run")
		return nil
	})
	defer SetExternalRunner(prevExt)

	_, err := ExecuteLine(context.Background(), nil, "login devhost", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try: %q", err.Error())
	}
}

func TestExecuteLine_UnknownWordNotSSH(t *testing.T) {
	cliroute.SetHostConfiguredHook(func(string) bool { return false })
	t.Cleanup(func() { cliroute.SetHostConfiguredHook(nil) })

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "nosuchcmd" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("external runner should not run")
		return nil
	})
	defer SetExternalRunner(prevExt)

	_, err := ExecuteLine(context.Background(), nil, "nosuchcmd", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestExecuteLine_SSHFlagLedConnect(t *testing.T) {
	cliroute.SetHostConfiguredHook(func(string) bool { return false })
	t.Cleanup(func() { cliroute.SetHostConfiguredHook(nil) })

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "-p" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	_, err := ExecuteLine(context.Background(), nil, "-p 22 newhost", nil)
	if err == nil {
		t.Fatal("expected connect error (not command not found)")
	}
	if errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("flag-led argv must attempt SSH connect, got %v", err)
	}
}

func TestExecuteLine_SSHStyleConnect(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	cliroute.SetHostConfiguredHook(func(name string) bool { return name == "devhost" })
	t.Cleanup(func() { cliroute.SetHostConfiguredHook(nil) })

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "devhost" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	_, err := ExecuteLine(context.Background(), nil, "devhost", nil)
	if err == nil {
		t.Fatal("expected connect error (not command not found)")
	}
	if errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("config host must attempt SSH connect, got %v", err)
	}
}

func TestExecuteLine_SSHHostPortConnect(t *testing.T) {
	cliroute.SetHostConfiguredHook(func(string) bool { return false })
	t.Cleanup(func() { cliroute.SetHostConfiguredHook(nil) })

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "192.168.1.1:2222" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	_, err := ExecuteLine(context.Background(), nil, "192.168.1.1:2222", nil)
	if err == nil {
		t.Fatal("expected connect error (not command not found)")
	}
	if errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("host:port must attempt SSH connect, got %v", err)
	}
}

func TestExecuteLine_SSHConnectBlockedWhenActive(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	cliroute.SetHostConfiguredHook(func(string) bool { return true })
	t.Cleanup(func() { cliroute.SetHostConfiguredHook(nil) })

	st, err := store.OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	sess := &store.Session{
		ID: "s1", Host: "h.example", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, Status: store.StatusActive,
	}
	if err := st.UpsertSession(sess); err != nil {
		t.Fatal(err)
	}
	if err := st.SetMeta(store.MetaCurrentSession, "s1"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Session.AuthTTL = "168h"
	mgr := &session.Manager{Store: st, Config: cfg}

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "devhost" for "%s"`, toolname.AISsh))
	}
	defer func() { cliExecutor = prevCLI }()

	_, err = ExecuteLine(context.Background(), mgr, "devhost", nil)
	if err == nil {
		t.Fatal("expected blocked error")
	}
	if !errors.Is(err, builtin.ErrLoginBlockedActive) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try: %q", err.Error())
	}
}
