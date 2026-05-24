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

package cli

import (
	"errors"
	"runtime"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
)

func wrapEditorErr(err error) error {
	if !errors.Is(err, platform.ErrNoEditor) {
		return err
	}
	if runtime.GOOS == "windows" {
		return errhint.WithTryMsg(
			platform.ErrNoEditor.Error(),
			"Set EDITOR or use a GUI editor.",
			"set EDITOR=notepad",
			"set EDITOR=code --wait",
		)
	}
	return errhint.WithTryMsg(
		platform.ErrNoEditor.Error(),
		"Set EDITOR or install vi.",
		"export EDITOR=vi",
		"export EDITOR=\"code --wait\"",
	)
}
