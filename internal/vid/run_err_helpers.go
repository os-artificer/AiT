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

package aivid

import (
	"errors"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
)

// printRunErr writes CLI errhint for err unless it is a partial-failure sentinel.
func printRunErr(w io.Writer, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, vidsearch.ErrPartialFailure) {
		return
	}
	errhint.Print(w, errhint.SurfaceCLI, err)
}
