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
	"errors"
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func runClear(stdout, stderr io.Writer, paths []string, f clearFlags) error {
	mask := flagsToMask(f)
	if mask.Empty() {
		return errhint.WithTry(
			imgmeta.ErrNothingToChange,
			"Specify at least one field flag or --all.",
			toolname.AIImg+" clear --all photo.jpg",
			toolname.AIImg+" clear --gps --make photo.jpg",
			toolname.AIImg+" clear --help",
		)
	}

	var anyFailed bool
	for _, path := range paths {
		if err := imgmeta.Clear(path, mask); err != nil {
			anyFailed = true
			errhint.Print(stderr, errhint.SurfaceCLI, wrapClearErr(path, err))
			continue
		}
		_, _ = fmt.Fprintf(stdout, "cleared %s\n", path)
	}
	if anyFailed {
		return imgmeta.ErrPartialFailure
	}
	return nil
}

func flagsToMask(f clearFlags) imgmeta.ClearMask {
	if f.all {
		return imgmeta.AllClearMask()
	}
	return imgmeta.ClearMask{
		Make:         f.make,
		Model:        f.model,
		Software:     f.software,
		LensMake:     f.lensMake,
		LensModel:    f.lensModel,
		SerialNumber: f.serial,
		DateTime:     f.datetime,
		GPS:          f.gps,
		Altitude:     f.altitude,
	}
}

func wrapClearErr(path string, err error) error {
	tryHelp := toolname.AIImg + " clear --help"
	if errors.Is(err, imgmeta.ErrUnsupported) {
		return errhint.WithTry(err,
			fmt.Sprintf("Cannot clear EXIF for %s (JPEG only in this version).", path),
			toolname.AIImg+" clear --all photo.jpg",
			tryHelp,
		)
	}
	return errhint.WithTry(err,
		fmt.Sprintf("Failed to clear %s.", path),
		toolname.AIImg+" clear --all photo.jpg",
		tryHelp,
	)
}
