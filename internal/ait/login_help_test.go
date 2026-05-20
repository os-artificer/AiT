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
)

func TestLoginHelpContainsFlags(t *testing.T) {
	if !loginHelpContains(
		"-p port",
		"-J destination",
		"-o key=value",
		"--alias name",
		"Not supported",
		"-L, -R, -D",
		"Examples:",
		"OpenSSH-compatible flags:",
		"AiT flags:",
	) {
		t.Fatal("login help missing expected sections")
	}
}

func TestWriteLoginHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteLoginHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "ait login [flags]") {
		t.Errorf("missing usage line: %q", out[:min(200, len(out))])
	}
}

func TestLoginHelpRequested(t *testing.T) {
	if !loginHelpRequested([]string{"-h"}) || !loginHelpRequested([]string{"--help"}) {
		t.Error("expected help requested")
	}
	if loginHelpRequested([]string{"host"}) || loginHelpRequested([]string{"-h", "host"}) {
		t.Error("unexpected help requested")
	}
}

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
	cmd := newLoginCmd()
	if !cmd.SilenceUsage || !cmd.SilenceErrors {
		t.Fatalf("login cmd SilenceUsage=%v SilenceErrors=%v", cmd.SilenceUsage, cmd.SilenceErrors)
	}
}

func TestLoginCmdHelpRunE(t *testing.T) {
	cmd := newLoginCmd()
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
