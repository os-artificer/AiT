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

package config

import (
	"os"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// EnsureFileForEdit creates or merges config.yaml when missing (same as EnsureFile(false)).
// Used by ait config view and edit before reading or opening the file.
func EnsureFileForEdit() (EnsureFileResult, error) {
	return EnsureFile(false)
}

// ReadRawFile returns the platform config path and raw config.yaml bytes.
// The file must already exist; callers use EnsureFileForEdit first when appropriate.
func ReadRawFile() (path string, content []byte, err error) {
	path, err = platform.ConfigPath()
	if err != nil {
		return "", nil, err
	}

	content, err = os.ReadFile(path)
	if err != nil {
		return path, nil, err
	}
	return path, content, nil
}
