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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/os-artificer/AiT/internal/aitimg/version"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/geocode"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

type runOptions struct {
	JSON      bool
	NoGeocode bool
	All       bool
}

func runFiles(ctx context.Context, stdout, stderr io.Writer, paths []string, opts runOptions) error {
	if len(paths) == 0 {
		return errhint.WithTry(
			imgmeta.ErrEmptyPath,
			"Provide one or more image files.",
			toolname.AIImg+" photo.jpg",
			toolname.AIImg+" --json photo.jpg shot.png",
			toolname.AIImg+" --help",
		)
	}

	geo := testGeocodeClient
	if geo == nil {
		geo = &geocode.Client{
			UserAgent: fmt.Sprintf("AiT-%s/%s (%s)", toolname.AIImg, version.Short(), projectURL),
		}
	}

	var (
		okInfos      []imgmeta.Info
		anyFailed    bool
		geocodeCalls int
	)

	for _, path := range paths {
		info, err := imgmeta.Extract(path)
		if err != nil {
			anyFailed = true
			errhint.Print(stderr, errhint.SurfaceCLI, wrapExtractErr(path, err))
			continue
		}

		if !opts.NoGeocode && info.HasGPS() {
			if geocodeCalls > 0 {
				if err := testGeocodeSleep(ctx); err != nil {
					return err
				}
			}
			addr, gerr := geo.Reverse(ctx, *info.Latitude, *info.Longitude)
			geocodeCalls++
			if gerr == nil {
				info.Address = addr
			}
		}

		okInfos = append(okInfos, info)
	}

	if opts.JSON {
		if err := imgmeta.FormatJSON(stdout, okInfos, opts.All); err != nil {
			return err
		}
	} else {
		if err := writeHumanBlocks(stdout, okInfos, opts.All); err != nil {
			return err
		}
	}

	if anyFailed {
		return imgmeta.ErrPartialFailure
	}
	return nil
}

func writeHumanBlocks(stdout io.Writer, infos []imgmeta.Info, allOut bool) error {
	for i, info := range infos {
		if i > 0 {
			if _, err := fmt.Fprintln(stdout); err != nil {
				return err
			}
		}
		if err := imgmeta.FormatHuman(stdout, info); err != nil {
			return err
		}
		if !allOut {
			continue
		}
		if _, err := fmt.Fprintln(stdout); err != nil {
			return err
		}
		if err := imgmeta.FormatFullHuman(stdout, info.Exif); err != nil {
			return err
		}
	}
	return nil
}

func wrapExtractErr(path string, err error) error {
	tryPhoto := toolname.AIImg + " photo.jpg"
	tryHelp := toolname.AIImg + " --help"

	if errors.Is(err, imgmeta.ErrUnsupported) {
		return errhint.WithTry(
			err,
			fmt.Sprintf("Cannot read image metadata from %s.", path),
			tryPhoto,
			tryHelp,
		)
	}
	if errors.Is(err, imgmeta.ErrEmptyPath) {
		return errhint.WithTry(
			err,
			"Provide one or more image files.",
			tryPhoto,
			toolname.AIImg+" --json photo.jpg shot.png",
			tryHelp,
		)
	}
	if errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return errhint.WithTry(
			err,
			fmt.Sprintf("File not found: %s.", path),
			tryPhoto,
			tryHelp,
		)
	}
	return errhint.WithTry(
		err,
		fmt.Sprintf("Failed to read %s.", path),
		tryPhoto,
		tryHelp,
	)
}

func defaultGeocodeSleep(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(geocodeMinInterval):
		return nil
	}
}
