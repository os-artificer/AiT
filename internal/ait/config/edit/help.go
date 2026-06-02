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

package edit

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	introMain = "Open config.yaml in $VISUAL or $EDITOR (default: vi on Unix, Notepad on Windows). Validates YAML after the editor exits."
	usageLine = "ait config edit"
	notes     = "  Editor order: $VISUAL, then $EDITOR, then the OS default.\n"
)

// WriteHelp prints ait config edit --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection("Notes:", notes); err != nil {
		return err
	}
	if err := b.WriteExamples("ait config edit"); err != nil {
		return err
	}
	return b.Finish("edit")
}
