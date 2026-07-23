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

package set

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aitimg set --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"Set writable EXIF fields on JPEG images (in place).",
		"Unsupported for PNG/TIFF/WebP in this version. Does not change image pixels.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AIImg + " set [flags] <file...>"); err != nil {
		return err
	}
	if err := b.WriteSection("Arguments", "  <file...>    one or more JPEG paths\n"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", ""+
		"  --make string         camera make\n"+
		"  --model string        camera model\n"+
		"  --software string     software tag\n"+
		"  --lens-make string    lens make\n"+
		"  --lens-model string   lens model\n"+
		"  --serial string       body serial number\n"+
		"  --datetime string     RFC3339 or YYYY:MM:DD HH:MM:SS\n"+
		"  --gps lat,lon         decimal degrees\n"+
		"  --altitude meters     altitude in meters\n"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		toolname.AIImg+" set --make Canon photo.jpg",
		toolname.AIImg+" set --gps 31.2304,121.4737 --altitude 12 photo.jpg",
		toolname.AIImg+` set --datetime "2020:06:07 01:30:00" photo.jpg`,
	); err != nil {
		return err
	}
	return b.Finish(Name)
}
