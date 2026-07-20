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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/aitimg/version"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/geocode"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"

	"github.com/spf13/cobra"
)

const (
	toolName           = "aitimg"
	geocodeMinInterval = time.Second
)

// testGeocodeClient, when non-nil, replaces the default Nominatim client (unit tests).
var testGeocodeClient *geocode.Client

// testGeocodeSleep, when non-nil, replaces the rate-limit sleep (unit tests).
var testGeocodeSleep = defaultGeocodeSleep

func defaultGeocodeSleep(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(geocodeMinInterval):
		return nil
	}
}

// SetTestGeocodeClient sets the geocode client used by Execute (tests only).
func SetTestGeocodeClient(c *geocode.Client) {
	testGeocodeClient = c
}

// SetTestGeocodeSleep sets the rate-limit sleep used between geocode calls (tests only).
// Pass nil to restore the default one-second sleep.
func SetTestGeocodeSleep(fn func(context.Context) error) {
	if fn == nil {
		testGeocodeSleep = defaultGeocodeSleep
		return
	}
	testGeocodeSleep = fn
}

// Execute runs the Cobra root with args (typically os.Args[1:]).
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the root command with injectable writers (tests).
func ExecuteWriters(args []string, stdout, stderr io.Writer) error {
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd.Execute()
}

func newRootCmd() *cobra.Command {
	var (
		jsonOut   bool
		noGeocode bool
		showVer   bool
	)

	cmd := &cobra.Command{
		Use:           toolName + " [flags] <file...>",
		Short:         "Extract image metadata (EXIF, GPS, device)",
		Long:          "Run 'aitimg --help' for usage and examples.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if showVer {
				_, err := fmt.Fprintln(c.OutOrStdout(), version.Version())
				return err
			}
			return runFiles(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), args, jsonOut, noGeocode)
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON array to stdout")
	cmd.Flags().BoolVar(&noGeocode, "no-geocode", false, "skip reverse geocode; Location shows 0")
	cmd.Flags().BoolVarP(&showVer, "version", "V", false, "print build version and exit")

	pkghelpdoc.Bind(cmd, WriteRootHelp)
	return cmd
}

func runFiles(ctx context.Context, stdout, stderr io.Writer, paths []string, jsonOut, noGeocode bool) error {
	if len(paths) == 0 {
		return errhint.WithTryMsg(
			"image path required",
			"Provide one or more image files.",
			"aitimg photo.jpg",
			"aitimg --json photo.jpg shot.png",
			"aitimg --help",
		)
	}

	geo := testGeocodeClient
	if geo == nil {
		geo = &geocode.Client{
			UserAgent: fmt.Sprintf("AiT-aitimg/%s (https://github.com/os-artificer/AiT)", version.Short()),
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

		if !noGeocode && info.HasGPS() {
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

	if jsonOut {
		if err := imgmeta.FormatJSON(stdout, okInfos); err != nil {
			return err
		}
	} else {
		for i, info := range okInfos {
			if i > 0 {
				if _, err := fmt.Fprintln(stdout); err != nil {
					return err
				}
			}
			if err := imgmeta.FormatHuman(stdout, info); err != nil {
				return err
			}
		}
	}

	if anyFailed {
		return errPartialFailure
	}
	return nil
}

var errPartialFailure = errors.New("one or more files failed")

func wrapExtractErr(path string, err error) error {
	if errors.Is(err, imgmeta.ErrUnsupported) {
		return errhint.WithTry(
			err,
			fmt.Sprintf("Cannot read image metadata from %s.", path),
			"aitimg photo.jpg",
			"aitimg --help",
		)
	}
	if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
		return errhint.WithTry(
			err,
			fmt.Sprintf("File not found: %s.", path),
			"aitimg photo.jpg",
			"aitimg --help",
		)
	}
	return errhint.WithTry(
		err,
		fmt.Sprintf("Failed to read %s.", path),
		"aitimg photo.jpg",
		"aitimg --help",
	)
}
