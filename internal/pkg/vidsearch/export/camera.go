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

package export

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
)

var (
	reCamDir = regexp.MustCompile(`(?i)^cam(?:era)?[-_]?(\d+)$`)
)

// ResolveCameraID picks camera id: override → env → path heuristic → unknown.
func ResolveCameraID(override, videoPath string) string {
	if s := strings.TrimSpace(override); s != "" {
		return s
	}
	if s := strings.TrimSpace(deps.CameraIDFromEnv()); s != "" {
		return s
	}
	dir := filepath.Base(filepath.Dir(videoPath))
	if m := reCamDir.FindStringSubmatch(dir); len(m) == 2 {
		return "cam" + m[1]
	}
	return "unknown"
}
