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
	"bytes"
	"strings"
	"testing"

	_ "github.com/os-artificer/AiT/internal/ait/registry/all"

	pkglogin "github.com/os-artificer/AiT/internal/ait/login"
)

func TestLoginCmdErrorNoUsage(t *testing.T) {
	root := newRootCmd()
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"login"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing destination")
	}
	if strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr must not contain Usage on error, got: %q", stderr.String())
	}
}

func TestLoginCmdSilenceUsage(t *testing.T) {
	cmd := pkglogin.New(testDeps().RunWithMgr)
	if !cmd.SilenceUsage || !cmd.SilenceErrors {
		t.Fatalf("login cmd SilenceUsage=%v SilenceErrors=%v", cmd.SilenceUsage, cmd.SilenceErrors)
	}
}

func TestLoginCmdHelpRunE(t *testing.T) {
	cmd := pkglogin.New(testDeps().RunWithMgr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.RunE(cmd, []string{"--help"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "-p port") {
		t.Fatalf("help output missing -p: %s", buf.String())
	}
}
