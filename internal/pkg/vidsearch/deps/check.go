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
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Need selects which dependencies to require.
type Need struct {
	Frames  bool // ffmpeg + ffprobe
	OCR     bool // tesseract soft unless RequiredOCR
	OCRHard bool
}

// Level is hard vs soft dependency.
type Level string

const (
	LevelHard Level = "hard"
	LevelSoft Level = "soft"
)

// DepName identifies a dependency binary.
type DepName string

const (
	DepFFmpeg    DepName = "ffmpeg"
	DepFFprobe   DepName = "ffprobe"
	DepTesseract DepName = "tesseract"
)

// Dep is one missing or present dependency.
type Dep struct {
	Name  DepName
	Level Level
	OK    bool
}

// Report is the result of Check.
type Report struct {
	Deps   []Dep
	Target Target
}

// MissingHard returns hard deps that are not OK.
func (r Report) MissingHard() []Dep {
	var out []Dep
	for _, d := range r.Deps {
		if d.Level == LevelHard && !d.OK {
			out = append(out, d)
		}
	}
	return out
}

// MissingSoft returns soft deps that are not OK.
func (r Report) MissingSoft() []Dep {
	var out []Dep
	for _, d := range r.Deps {
		if d.Level == LevelSoft && !d.OK {
			out = append(out, d)
		}
	}
	return out
}

// HasFFmpeg reports ffmpeg and ffprobe present.
func (r Report) HasFFmpeg() bool {
	okF, okP := false, false
	for _, d := range r.Deps {
		if d.Name == DepFFmpeg {
			okF = d.OK
		}
		if d.Name == DepFFprobe {
			okP = d.OK
		}
	}
	return okF && okP
}

// HasTesseract reports tesseract present.
func (r Report) HasTesseract() bool {
	for _, d := range r.Deps {
		if d.Name == DepTesseract {
			return d.OK
		}
	}
	return false
}

// ErrMissingFFmpeg is returned when ffmpeg/ffprobe are missing.
var ErrMissingFFmpeg = errhint.NewSentinel("ffmpeg or ffprobe not found")

// ErrMissingTesseract is returned when tesseract is required but missing.
var ErrMissingTesseract = errhint.NewSentinel("tesseract not found")

// ErrInstallNeedsYes is returned for non-TTY install without --yes.
var ErrInstallNeedsYes = errhint.NewSentinel("install requires confirmation")

// Check probes PATH for required tools.
func Check(need Need, lookPath LookPathFunc) Report {
	if lookPath == nil {
		lookPath = DefaultLookPath
	}
	target := DetectTarget(lookPath)
	rep := Report{Target: target}

	if need.Frames {
		rep.Deps = append(rep.Deps, probe(lookPath, DepFFmpeg, LevelHard))
		rep.Deps = append(rep.Deps, probe(lookPath, DepFFprobe, LevelHard))
	}
	if need.OCR || need.OCRHard {
		level := LevelSoft
		if need.OCRHard {
			level = LevelHard
		}
		rep.Deps = append(rep.Deps, probe(lookPath, DepTesseract, level))
	}
	return rep
}

func probe(lookPath LookPathFunc, name DepName, level Level) Dep {
	_, err := lookPath(string(name))
	return Dep{Name: name, Level: level, OK: err == nil}
}

// MissingErr builds errhint for hard missing deps.
// Prefer ffmpeg/ffprobe over tesseract when both are missing (frames are prerequisite).
func MissingErr(rep Report) error {
	hard := rep.MissingHard()
	if len(hard) == 0 {
		return nil
	}
	for _, d := range hard {
		if d.Name == DepFFmpeg || d.Name == DepFFprobe {
			return ffmpegMissingErr(rep.Target)
		}
	}
	for _, d := range hard {
		if d.Name == DepTesseract {
			return tesseractMissingErr(rep.Target)
		}
	}
	return ffmpegMissingErr(rep.Target)
}

func ffmpegMissingErr(target Target) error {
	tries := []string{
		toolname.VidCLI("deps --install"),
		toolname.AITCLI("vid deps --install"),
	}
	if cmd := InstallFFmpegCommand(target); cmd != "" {
		tries = append(tries, cmd)
	} else {
		tries = append(tries, "see "+ManualFFmpegURL)
	}
	return errhint.WithTry(ErrMissingFFmpeg,
		"Video search needs ffmpeg and ffprobe on PATH.",
		tries...,
	)
}

func tesseractMissingErr(target Target) error {
	tries := []string{
		toolname.VidCLI("deps --install"),
		toolname.AITCLI("vid deps --install"),
	}
	if cmd := InstallTesseractCommand(target); cmd != "" {
		tries = append(tries, cmd)
	} else {
		tries = append(tries, "see "+ManualTesseractURL)
	}
	return errhint.WithTry(ErrMissingTesseract,
		"Local OCR requires tesseract on PATH (or disable --local-only and use a vision model).",
		tries...,
	)
}
