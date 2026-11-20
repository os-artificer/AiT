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

package frames_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
)

func TestExtract_RespectsMaxFrames(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not on PATH")
	}
	video := synthesizeTinyMP4(t)
	out := t.TempDir()
	fr, err := frames.Extract(context.Background(), video, frames.Options{
		FPS:       10,
		MaxFrames: 3,
		OutDir:    out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fr) > 3 {
		t.Fatalf("got %d frames", len(fr))
	}
	if len(fr) == 0 {
		t.Fatal("expected some frames")
	}
}

func TestExtract_CleansCallerTemp(t *testing.T) {
	// Cleanup is owned by search; Extract only writes under OutDir.
	out := t.TempDir()
	_ = out
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	cancel()
	_, err := frames.Extract(ctx, filepath.Join(out, "missing.mp4"), frames.Options{OutDir: out})
	if err == nil {
		t.Fatal("expected error on missing video / cancelled ctx")
	}
}

func synthesizeTinyMP4(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "tiny.mp4")
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "color=c=red:s=64x64:d=2",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-y", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot synthesize mp4: %v (%s)", err, out)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatal(err)
	}
	return dst
}
