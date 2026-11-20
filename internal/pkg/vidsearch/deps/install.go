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

package deps

import (
	"context"
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// InstallOpts controls automatic dependency installation.
type InstallOpts struct {
	Yes      bool
	IsTTY    bool
	Confirm  func(prompt string) (bool, error) // optional; used when IsTTY && !Yes
	NeedOCR  bool
	Runner   CommandRunner
	LookPath LookPathFunc
}

// Install installs missing ffmpeg (always when missing) and optionally tesseract.
func Install(ctx context.Context, opts InstallOpts) error {
	runner := opts.Runner
	if runner == nil {
		runner = DefaultRunner()
	}
	look := opts.LookPath
	if look == nil {
		look = runner.LookPath
	}

	need := Need{Frames: true, OCR: opts.NeedOCR}
	rep := Check(need, look)
	missing := append(rep.MissingHard(), rep.MissingSoft()...)
	if len(missing) == 0 {
		return nil
	}

	if !opts.Yes {
		if !opts.IsTTY {
			return errhint.WithTry(ErrInstallNeedsYes,
				"Non-interactive install requires --yes.",
				toolname.VidCLI("deps --install --yes"),
			)
		}
		if opts.Confirm != nil {
			ok, err := opts.Confirm("Install missing video dependencies via the system package manager? [y/N] ")
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("install declined")
			}
		}
	}

	target := rep.Target
	wantFF, wantTess := false, false
	for _, d := range missing {
		switch d.Name {
		case DepFFmpeg, DepFFprobe:
			wantFF = true
		case DepTesseract:
			wantTess = true
		}
	}

	if wantFF {
		if err := runInstallLine(ctx, runner, InstallFFmpegCommand(target), target, "ffmpeg"); err != nil {
			return err
		}
	}
	if wantTess {
		if err := runInstallLine(ctx, runner, InstallTesseractCommand(target), target, "tesseract"); err != nil {
			return err
		}
	}

	after := Check(need, look)
	if err := MissingErr(after); err != nil {
		return err
	}
	if wantTess && !after.HasTesseract() {
		return tesseractMissingErr(target)
	}
	return nil
}

func runInstallLine(ctx context.Context, runner CommandRunner, line string, target Target, label string) error {
	if line == "" {
		url := ManualFFmpegURL
		if label == "tesseract" {
			url = ManualTesseractURL
		}
		return fmt.Errorf("no package manager install command for %s; see %s", label, url)
	}
	name, args := splitShell(line)
	if name == "" {
		return fmt.Errorf("empty install command for %s", label)
	}
	// Prefer sh -c for compound apt lines.
	if strings.Contains(line, "&&") || strings.HasPrefix(line, "sudo ") {
		return runner.Run(ctx, "sh", []string{"-c", line})
	}
	return runner.Run(ctx, name, args)
}

func splitShell(line string) (string, []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}

// FormatStatus writes a human-readable dependency status report.
func FormatStatus(rep Report) string {
	var b strings.Builder
	for _, d := range rep.Deps {
		status := "missing"
		if d.OK {
			status = "ok"
		}
		fmt.Fprintf(&b, "%s  %s  (%s)\n", status, d.Name, d.Level)
		if !d.OK {
			var cmd string
			switch d.Name {
			case DepFFmpeg, DepFFprobe:
				cmd = InstallFFmpegCommand(rep.Target)
			case DepTesseract:
				cmd = InstallTesseractCommand(rep.Target)
			}
			if cmd != "" {
				fmt.Fprintf(&b, "  try: %s\n", cmd)
			}
			fmt.Fprintf(&b, "  try: %s\n", toolname.VidCLI("deps --install"))
		}
	}
	return b.String()
}
