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

// Package aissh implements the AiT aissh CLI application entry.
package aissh

import (
	"os"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/ssh/connect"
)

// Run is the application entry invoked from cmd/aissh.
// args are os.Args[1:] passed to the Cobra root command.
// Returns Cobra RunE errors (connect, session, get, put, etc.) printed to stderr before return.
func Run(args []string) error {
	if err := Execute(args); err != nil {
		errhint.Print(os.Stderr, errhint.SurfaceCLI, err)
		return err
	}
	return nil
}

// ExitCode returns the process exit code if err is from a remote command.
func ExitCode(err error) (int, bool) {
	return connect.ExitCode(err)
}
