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

package stop

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain    = "Stop local ollama serve processes. Does not disable Ollama in config."
	introWindows = "On Windows, may not quit the Ollama desktop app; close it manually if needed."
)

var (
	usageLine   = toolname.AICLI("ollama stop")
	exampleStop = toolname.AICLI("ollama stop")
)

// WriteHelp prints ai ollama stop --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introWindows); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleStop); err != nil {
		return err
	}
	return b.Finish("stop")
}
