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

package clear

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aiimg clear --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"Clear writable EXIF fields on JPEG images.",
		"By default writes a new sibling file (*_aiimg.jpg). On a TTY, asks whether to overwrite originals. Use --all or field flags. JPEG only.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AIImg + " clear [flags] <file...>"); err != nil {
		return err
	}
	if err := b.WriteSection("Arguments", "  <file...>    one or more JPEG paths\n"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", ""+
		"  --make          clear camera make\n"+
		"  --model         clear camera model\n"+
		"  --software      clear software tag\n"+
		"  --lens-make     clear lens make\n"+
		"  --lens-model    clear lens model\n"+
		"  --serial        clear body serial number\n"+
		"  --datetime      clear capture time\n"+
		"  --gps           clear GPS latitude/longitude\n"+
		"  --altitude      clear altitude\n"+
		"  --all           clear all writable fields\n"+
		"  -o, --output path  write to path (single input; no prompt)\n"+
		"  --inplace       overwrite original file(s)\n"+
		"  -y, --yes       same as --inplace\n"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		toolname.AIImg+" clear --all photo.jpg",
		toolname.AIImg+" clear --all --inplace photo.jpg",
		toolname.AIImg+" clear --gps --make -o out.jpg photo.jpg",
	); err != nil {
		return err
	}
	return b.Finish(Name)
}
