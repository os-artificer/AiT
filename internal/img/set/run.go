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
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/os-artificer/AiT/internal/img/outpath"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func runSet(stdout, stderr io.Writer, paths []string, f setFlags) error {
	patch, err := flagsToPatch(f)
	if err != nil {
		return err
	}
	if patch.Empty() {
		return errhint.WithTry(
			imgmeta.ErrNothingToChange,
			"Specify at least one field flag to set.",
			toolname.AIImg+" set --make Canon photo.jpg",
			toolname.AIImg+" set --help",
		)
	}

	dsts, err := outpath.Resolve(outpath.ResolveOpts{
		Paths:     paths,
		Output:    f.output,
		OutputSet: f.outputSet,
		Inplace:   f.inplace || f.yes,
		PromptOut: stderr,
		ConfirmIn: os.Stdin,
	})
	if err != nil {
		return err
	}

	var anyFailed bool
	for i, src := range paths {
		dst := dsts[i]
		if err := imgmeta.PatchTo(src, dst, patch); err != nil {
			anyFailed = true
			errhint.Print(stderr, errhint.SurfaceCLI, wrapSetErr(src, err))
			continue
		}
		outpath.PrintResult(stdout, "updated", src, dst)
	}
	if anyFailed {
		return imgmeta.ErrPartialFailure
	}
	return nil
}

func flagsToPatch(f setFlags) (imgmeta.PatchFields, error) {
	var p imgmeta.PatchFields
	if f.sawMake {
		v := f.make
		p.Make = &v
	}
	if f.sawModel {
		v := f.model
		p.Model = &v
	}
	if f.sawSoftware {
		v := f.software
		p.Software = &v
	}
	if f.sawLensMake {
		v := f.lensMake
		p.LensMake = &v
	}
	if f.sawLensModel {
		v := f.lensModel
		p.LensModel = &v
	}
	if f.sawSerial {
		v := f.serial
		p.SerialNumber = &v
	}
	if f.sawDatetime {
		t, err := imgmeta.ParseDateTime(f.datetime)
		if err != nil {
			return p, errhint.WithTry(err, "Invalid --datetime value.",
				toolname.AIImg+` set --datetime "2020:06:07 01:30:00" photo.jpg`,
				toolname.AIImg+" set --help",
			)
		}
		p.DateTime = &t
	}
	if f.sawGPS {
		lat, lon, err := parseGPS(f.gps)
		if err != nil {
			return p, errhint.WithTry(err, "Invalid --gps value.",
				toolname.AIImg+" set --gps 31.2304,121.4737 photo.jpg",
				toolname.AIImg+" set --help",
			)
		}
		p.Latitude, p.Longitude = &lat, &lon
	}
	if f.sawAltitude {
		alt, err := strconv.ParseFloat(f.altitude, 64)
		if err != nil {
			return p, errhint.WithTry(err, "Invalid --altitude value.",
				toolname.AIImg+" set --altitude 12.5 photo.jpg",
				toolname.AIImg+" set --help",
			)
		}
		p.Altitude = &alt
	}
	return p, nil
}

func parseGPS(s string) (lat, lon float64, err error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("want lat,lon")
	}
	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, err
	}
	lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, nil
}

func wrapSetErr(path string, err error) error {
	tryHelp := toolname.AIImg + " set --help"
	if errors.Is(err, imgmeta.ErrUnsupported) {
		return errhint.WithTry(err,
			fmt.Sprintf("Cannot write EXIF for %s (JPEG only in this version).", path),
			toolname.AIImg+" set --make Canon photo.jpg",
			tryHelp,
		)
	}
	return errhint.WithTry(err,
		fmt.Sprintf("Failed to update %s.", path),
		toolname.AIImg+" set --make Canon photo.jpg",
		tryHelp,
	)
}
