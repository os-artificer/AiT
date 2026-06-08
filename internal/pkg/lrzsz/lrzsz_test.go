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

package lrzsz

import (
	"strings"
	"testing"
)

func TestParseProbeOK(t *testing.T) {
	if !ParseProbeOK("ok") {
		t.Fatal("expected ok")
	}
	if !ParseProbeOK("noise\nok\n") {
		t.Fatal("expected ok with noise")
	}
	if ParseProbeOK("missing") {
		t.Fatal("expected false")
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		out  string
		want Target
	}{
		{"linux-apt", TargetLinuxApt},
		{"linux-dnf", TargetLinuxDnf},
		{"darwin-brew", TargetDarwinBrew},
		{"windows-choco", TargetWindowsChoco},
		{"noise\nlinux-apk\n", TargetLinuxApk},
		{"", TargetUnknown},
	}
	for _, tc := range tests {
		got := ParseTarget(tc.out)
		if got != tc.want {
			t.Errorf("ParseTarget(%q) = %q, want %q", tc.out, got, tc.want)
		}
	}
}

func TestInstallCommand_LinuxApt(t *testing.T) {
	cmd := InstallCommand(TargetLinuxApt)
	if !strings.Contains(cmd, "apt-get install") || !strings.Contains(cmd, "lrzsz") {
		t.Fatalf("unexpected apt cmd: %q", cmd)
	}
}

func TestInstallCommand_DarwinBrew(t *testing.T) {
	cmd := InstallCommand(TargetDarwinBrew)
	if cmd != "brew install lrzsz" {
		t.Fatalf("got %q", cmd)
	}
}

func TestInstallCommand_UnknownEmpty(t *testing.T) {
	if InstallCommand(TargetUnknown) != "" {
		t.Fatal("expected empty for unknown")
	}
	if InstallCommand(TargetDarwinNoBrew) != "" {
		t.Fatal("expected empty for darwin-no-brew")
	}
}

func TestManualTryLines_DarwinNoBrew(t *testing.T) {
	lines := ManualTryLines(TargetDarwinNoBrew)
	if len(lines) < 2 || !strings.Contains(lines[0], "brew install lrzsz") {
		t.Fatalf("unexpected lines: %v", lines)
	}
}

func TestWrapMissingChain_Format(t *testing.T) {
	err := WrapMissingChain("internal", TargetLinuxApt)
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try: %q", s)
	}
	if !strings.Contains(s, "apt-get install") {
		t.Fatalf("missing apt hint: %q", s)
	}
}

func TestResolveInstallAllowed_AssumeYes(t *testing.T) {
	ok, err := ResolveInstallAllowed(true, InstallOptions{AssumeYes: true}, nil)
	if err != nil || !ok {
		t.Fatalf("AssumeYes: ok=%v err=%v", ok, err)
	}
}

func TestResolveInstallAllowed_NoInstall(t *testing.T) {
	ok, err := ResolveInstallAllowed(true, InstallOptions{NoInstall: true}, nil)
	if err != nil || ok {
		t.Fatalf("NoInstall: ok=%v err=%v", ok, err)
	}
}

func TestResolveInstallAllowed_NonTTY(t *testing.T) {
	_, err := ResolveInstallAllowed(true, InstallOptions{In: strings.NewReader("")}, nil)
	if err != ErrNonInteractiveInstall {
		t.Fatalf("got %v", err)
	}
}

// TestInstallPrompt_ActiveShellHostLabel (C22) keeps the empty-host install label unchanged.
func TestInstallPrompt_ActiveShellHostLabel(t *testing.T) {
	got := InstallPrompt("")
	if !strings.Contains(got, labelActiveShellHost) {
		t.Fatalf("prompt=%q", got)
	}
}
