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

package ait_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease/testrelease"
)

func TestExecuteLine_InstallUnknownPlugin(t *testing.T) {
	_, err := ait.ExecuteLine(context.Background(), "install nosuch", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, aitrelease.ErrUnknownPlugin) {
		t.Fatalf("got %v", err)
	}
}

func TestExecuteLine_UpgradeRejectsArgs(t *testing.T) {
	_, err := ait.ExecuteLine(context.Background(), "upgrade aissh", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, aitrelease.ErrUpgradeExtraArgs) {
		t.Fatalf("got %v", err)
	}
}

func TestExecuteLine_UpdateNothingInstalled(t *testing.T) {
	resetPlugins(t)
	_, err := ait.ExecuteLine(context.Background(), "update", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, aitrelease.ErrNothingToUpdate) {
		t.Fatalf("got %v", err)
	}
}

func TestWriteRootHelp_IncludesManageCmds(t *testing.T) {
	var buf bytes.Buffer
	if err := ait.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, cmd := range []string{"install", "update", "upgrade"} {
		if !strings.Contains(s, cmd) {
			t.Fatalf("missing %q in help: %s", cmd, s)
		}
	}
}

func TestWriteManageHelp_Install(t *testing.T) {
	var buf bytes.Buffer
	exit, err := ait.ExecuteLine(context.Background(), "help install", nil, &buf, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if exit {
		t.Fatal("unexpected exit")
	}
	if !strings.Contains(buf.String(), "Usage:") || !strings.Contains(buf.String(), "install") {
		t.Fatalf("help=%q", buf.String())
	}
}

func TestExecuteLine_UpdateSkipsMissing(t *testing.T) {
	installBuiltinPlugin(t, "aissh")
	srv := testrelease.NewServer(t, "v0.9.0")
	t.Cleanup(srv.Close)
	t.Setenv(aitrelease.TestReleaseBaseURLEnv, srv.URL)

	var out bytes.Buffer
	_, err := ait.ExecuteLine(context.Background(), "update aissh ailog", nil, &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	body := out.String()
	if !strings.Contains(body, "skipped ailog") {
		t.Fatalf("out=%q", body)
	}
	if !strings.Contains(body, "updated aissh") {
		t.Fatalf("out=%q", body)
	}
}

func TestExecuteLine_UpgradeIncludesAIT(t *testing.T) {
	installBuiltinPlugin(t, "aissh")
	srv := testrelease.NewServer(t, "v0.9.0")
	t.Cleanup(srv.Close)
	t.Setenv(aitrelease.TestReleaseBaseURLEnv, srv.URL)

	var out bytes.Buffer
	_, err := ait.ExecuteLine(context.Background(), "upgrade", nil, &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "upgrade complete") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestExecuteLine_InstallOneIntegration(t *testing.T) {
	root := t.TempDir()
	aitplugin.SetInstallDirsForTest([]string{root})
	resetPlugins(t)

	srv := testrelease.NewServer(t, "v0.9.0")
	t.Cleanup(srv.Close)
	t.Setenv(aitrelease.TestReleaseBaseURLEnv, srv.URL)

	var out bytes.Buffer
	_, err := ait.ExecuteLine(context.Background(), "install aissh", nil, &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "installed aissh") {
		t.Fatalf("out=%q", out.String())
	}
}
