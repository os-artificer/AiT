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
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
)

func TestRun_PartialFailure_NoErrhint(t *testing.T) {
	var buf bytes.Buffer
	err := errhint.WithTry(vidsearch.ErrPartialFailure, "some videos failed", "aitvid search ./clip.mp4 --query x")
	printRunErr(&buf, err)
	if buf.Len() != 0 {
		t.Fatalf("stderr=%q", buf.String())
	}
}

func TestRun_MissingFFmpeg_Errhint(t *testing.T) {
	var buf bytes.Buffer
	err := errhint.WithTry(deps.ErrMissingFFmpeg, "Install ffmpeg.", "aitvid deps --install")
	printRunErr(&buf, fmt.Errorf("wrap: %w", err))
	s := buf.String()
	if !strings.Contains(s, "Try:") && !strings.Contains(s, "deps --install") {
		t.Fatalf("stderr=%q", s)
	}
}
