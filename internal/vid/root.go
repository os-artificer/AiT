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

// Package aivid implements the AiT aitvid CLI application entry.
package aivid

import (
	"errors"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
)

// Run is the application entry invoked from cmd/vid.
func Run(args []string) error {
	if err := Execute(args); err != nil {
		if !errors.Is(err, vidsearch.ErrPartialFailure) {
			errhint.Print(os.Stderr, errhint.SurfaceCLI, err)
		}
		return err
	}
	return nil
}

// ExitCode maps errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	return 1
}
