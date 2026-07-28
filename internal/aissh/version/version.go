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

package version

import (
	_ "embed"

	"github.com/os-artificer/AiT/internal/pkg/buildinfo"
)

//go:embed VERSION
var embeddedVersion string

// Build metadata injected via -ldflags at make build/install time.
var (
	version   = ""
	gitCommit = ""
	buildDate = ""
)

func inputs() buildinfo.Inputs {
	return buildinfo.Inputs{
		Version:   version,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		Embedded:  embeddedVersion,
	}
}

// Version returns a human-readable build version string.
func Version() string {
	return buildinfo.Format(inputs())
}

// Short returns a compact version token for User-Agent.
func Short() string {
	return buildinfo.Short(inputs())
}
