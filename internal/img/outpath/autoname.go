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

package outpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const autoNameTag = "_aiimg"

// AutoName returns a sibling path for src: stem_aiimg.ext, then stem_aiimg_1.ext, …
// The chosen path does not exist at the time of the check (TOCTOU accepted).
func AutoName(src string) (string, error) {
	if src == "" {
		return "", fmt.Errorf("empty source path")
	}
	dir := filepath.Dir(src)
	base := filepath.Base(src)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	candidate := filepath.Join(dir, stem+autoNameTag+ext)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate, nil
	}
	for i := 1; i < 10000; i++ {
		candidate = filepath.Join(dir, stem+autoNameTag+"_"+strconv.Itoa(i)+ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate auto output name for %s", src)
}
