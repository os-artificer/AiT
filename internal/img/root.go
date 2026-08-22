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

// Package aiimg implements the AiT aiimg CLI application entry.
package aiimg

import (
	"errors"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

// Run is the application entry invoked from cmd/aiimg.
// args are os.Args[1:] passed to Execute.
func Run(args []string) error {
	if err := Execute(args); err != nil {
		if !errors.Is(err, imgmeta.ErrPartialFailure) {
			errhint.Print(os.Stderr, errhint.SurfaceCLI, err)
		}
		return err
	}
	return nil
}
