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

package localmode

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

func TestEnv(t *testing.T) {
	env := Env([]string{"FOO=bar"})
	var hasMode, hasPS1, hasPrompt bool
	for _, e := range env {
		switch {
		case e == "AIT_LOCAL_MODE=1":
			hasMode = true
		case strings.HasPrefix(e, "PS1="):
			hasPS1 = true
			if !strings.Contains(e, "AiT Local>") {
				t.Errorf("PS1 should contain display prompt: %q", e)
			}
		case strings.HasPrefix(e, "PROMPT="):
			hasPrompt = true
			if !strings.Contains(e, prompt.Local) {
				t.Errorf("PROMPT should contain Local prompt: %q", e)
			}
		}
	}
	if !hasMode || !hasPS1 || !hasPrompt {
		t.Errorf("env missing vars: mode=%v ps1=%v prompt=%v", hasMode, hasPS1, hasPrompt)
	}
}

func TestPrepareBash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX init only")
	}
	init, args, extraEnv, err := Prepare("/bin/bash")
	if err != nil {
		t.Fatal(err)
	}
	defer init.remove()
	if init == nil || init.dir == "" {
		t.Fatal("expected init dir for bash")
	}
	if len(extraEnv) != 0 {
		t.Errorf("bash should not add extra env, got %v", extraEnv)
	}
	if len(args) != 2 || args[0] != "--rcfile" {
		t.Fatalf("want --rcfile, got %v", args)
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "export PS1=") {
		t.Errorf("rc should set PS1: %q", body)
	}
	if !strings.Contains(body, "AiT Local>") {
		t.Errorf("rc should contain prompt: %q", body)
	}
}

func TestPrepareZsh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX init only")
	}
	init, args, extraEnv, err := Prepare("/usr/bin/zsh")
	if err != nil {
		t.Fatal(err)
	}
	defer init.remove()
	if init == nil {
		t.Fatal("expected init for zsh")
	}
	if len(args) != 0 {
		t.Errorf("zsh should not use argv, got %v", args)
	}
	var zdot string
	for _, e := range extraEnv {
		if strings.HasPrefix(e, "ZDOTDIR=") {
			zdot = strings.TrimPrefix(e, "ZDOTDIR=")
		}
	}
	if zdot != init.dir {
		t.Errorf("ZDOTDIR=%q want %q", zdot, init.dir)
	}
	data, err := os.ReadFile(filepath.Join(init.dir, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "export PROMPT=") {
		t.Errorf("zshrc should set PROMPT: %q", body)
	}
	if strings.Count(body, "export PROMPT=") < 2 {
		t.Errorf("zshrc should re-apply PROMPT after user rc: %q", body)
	}
}

func TestPrepareOther(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX init only")
	}
	init, args, extraEnv, err := Prepare("/bin/fish")
	if err != nil {
		t.Fatal(err)
	}
	if init != nil {
		init.remove()
	}
	if init != nil || len(args) != 0 || len(extraEnv) != 0 {
		t.Errorf("fish should use env only: init=%v args=%v env=%v", init, args, extraEnv)
	}
}
