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

package aitimg

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// WriteRootHelp prints aitimg --help to w.
func WriteRootHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"Extract image metadata: device, capture time, GPS coordinates, and reverse-geocoded location.",
		"Formats: JPEG, PNG (TIFF/WebP best-effort). Address lookup uses OpenStreetMap Nominatim (needs network).",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("aitimg [flags] <file...>"); err != nil {
		return err
	}
	if err := b.WriteSection("Arguments", "  <file...>    one or more image paths\n"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", ""+
		"  --json           emit JSON array to stdout\n"+
		"  --no-geocode     skip reverse geocode; Location shows 0 (GPS still shown)\n"+
		"  -V, --version    print build version and exit\n"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"aitimg photo.jpg",
		"aitimg --json photo.jpg shot.png",
		"aitimg --no-geocode photo.jpg",
		"aitimg --version",
	); err != nil {
		return err
	}
	return b.Finish("aitimg")
}
