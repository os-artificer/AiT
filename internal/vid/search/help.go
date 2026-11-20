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

package search

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aitvid search --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"Search a video file or directory for text/objects via OCR (and optional vision).",
		"Hit keyframes are saved by default under --out-dir with sidecar JSON and pixel annotation.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.VidCLI(Name + " <path> --query <text> [flags]")); err != nil {
		return err
	}
	flags := "" +
		"  --query string        search query (required)\n" +
		"  --mode string         text|object|auto (default auto)\n" +
		"  --local-only          OCR only\n" +
		"  --fps float           sample rate (default 0.5)\n" +
		"  --max-frames int      per video (default 30)\n" +
		"  --max-videos int      directory cap (0 = unlimited, default)\n" +
		"  -r, --recursive       recurse directories\n" +
		"  --ext string          extension whitelist\n" +
		"  --save-hits           save hit frames (default true)\n" +
		"  --no-save-hits        disable save\n" +
		"  -o, --out-dir string  output dir (default ./aitvid-hits)\n" +
		"  --camera-id string    camera id override\n" +
		"  --annotate            burn provenance (default true)\n" +
		"  --json                NDJSON hits\n"
	if err := b.WriteSection(pkghelpdoc.SectionFlags, flags); err != nil {
		return err
	}
	return b.Finish(Name)
}
