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
	"io"

	"github.com/os-artificer/AiT/internal/pkg/geocode"
)

// testGeocodeClient, when non-nil, replaces the default Nominatim client (unit tests).
var testGeocodeClient *geocode.Client

// testGeocodeSleep, when non-nil, replaces the rate-limit sleep (unit tests).
var testGeocodeSleep = defaultGeocodeSleep

// Run parses argv and extracts metadata for the given image paths.
// ctx is attached to the private Cobra command via SetContext.
// Returns ErrPartialFailure when some files fail while others succeed.
func Run(ctx context.Context, stdout, stderr io.Writer, args []string) error {
	cmd := newCmd()
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd.Execute()
}

// SetTestGeocodeClient sets the geocode client used by Run (tests only).
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
