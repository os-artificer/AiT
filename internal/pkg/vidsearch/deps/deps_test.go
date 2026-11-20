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

package deps_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
)

func TestCheck_MissingFFmpeg(t *testing.T) {
	look := func(name string) (string, error) {
		return "", errors.New("not found")
	}
	rep := deps.Check(deps.Need{Frames: true}, look)
	if rep.HasFFmpeg() {
		t.Fatal("expected missing ffmpeg")
	}
	err := deps.MissingErr(rep)
	if !errors.Is(err, deps.ErrMissingFFmpeg) {
		t.Fatalf("got %v", err)
	}
	s := err.Error()
	if !strings.Contains(s, "deps --install") {
		t.Fatalf("missing install try: %s", s)
	}
	if !strings.Contains(s, toolname.AITVid) && !strings.Contains(s, toolname.Vid) {
		t.Fatalf("missing vid CLI: %s", s)
	}
}

func TestInstall_NeedsYesNonTTY(t *testing.T) {
	look := func(string) (string, error) { return "", errors.New("x") }
	err := deps.Install(context.Background(), deps.InstallOpts{
		Yes:      false,
		IsTTY:    false,
		LookPath: look,
		Runner:   &fakeRunner{look: look},
	})
	if !errors.Is(err, deps.ErrInstallNeedsYes) {
		t.Fatalf("got %v", err)
	}
}

func TestInstall_FakeRunner(t *testing.T) {
	// Seed package managers so DetectTarget is not empty on CI hosts.
	present := map[string]bool{"brew": true, "apt-get": true}
	look := func(name string) (string, error) {
		if present[name] {
			return "/bin/" + name, nil
		}
		return "", errors.New("missing")
	}
	fr := &fakeRunner{look: look, onRun: func(name string, args []string) {
		present["ffmpeg"] = true
		present["ffprobe"] = true
	}}
	err := deps.Install(context.Background(), deps.InstallOpts{
		Yes:      true,
		LookPath: look,
		Runner:   fr,
	})
	if err != nil {
		t.Fatal(err)
	}
	if fr.calls == 0 {
		t.Fatal("expected install run")
	}
}

func TestInstallCommand_DarwinBrew(t *testing.T) {
	cmd := deps.InstallFFmpegCommand(deps.TargetDarwinBrew)
	if cmd != "brew install ffmpeg" {
		t.Fatalf("got %q", cmd)
	}
}

type fakeRunner struct {
	look  deps.LookPathFunc
	onRun func(name string, args []string)
	calls int
}

func (f *fakeRunner) LookPath(name string) (string, error) {
	return f.look(name)
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string) error {
	f.calls++
	if f.onRun != nil {
		f.onRun(name, args)
	}
	return nil
}
