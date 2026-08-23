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

package extract

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain = "Extract image metadata: device, capture time, GPS coordinates, and reverse-geocoded location."
	introHint = "Formats: JPEG, PNG (TIFF/WebP best-effort). Address lookup uses OpenStreetMap Nominatim (needs network)."
	argsBody  = "  <file...>    one or more image paths\n"
	flagsBody = "" +
		"  --json           emit JSON array to stdout\n" +
		"  --all            print full structured EXIF (IFD0/ExifIFD/GPS/…) after summary\n" +
		"  --no-geocode     skip reverse geocode; Location shows 0 (GPS still shown)\n" +
		"  -V, --version    print build version and exit\n"
)

// WriteIntroUsageArgsFlags writes the default-route Intro through Flags sections into b.
// Shared by WriteHelp and the parent root help composer so Flags text stays single-sourced.
func WriteIntroUsageArgsFlags(b *pkghelpdoc.Body) error {
	if err := b.WriteIntro(introMain, introHint); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AIImg + " [flags] <file...>"); err != nil {
		return err
	}
	if err := b.WriteSection("Arguments", argsBody); err != nil {
		return err
	}
	return b.WriteSection("Flags", flagsBody)
}

// DefaultExamples returns default-route example lines (without the version subcommand).
func DefaultExamples() []string {
	return []string{
		toolname.AIImg + " photo.jpg",
		toolname.AIImg + " --json photo.jpg shot.png",
		toolname.AIImg + " --all photo.jpg",
		toolname.AIImg + " --json --all photo.jpg",
		toolname.AIImg + " --no-geocode photo.jpg",
		toolname.AIImg + " --version",
	}
}

// WriteHelp prints default-route usage (no Commands section) to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := WriteIntroUsageArgsFlags(b); err != nil {
		return err
	}
	if err := b.WriteExamples(DefaultExamples()...); err != nil {
		return err
	}
	return b.Finish(toolname.AIImg)
}
