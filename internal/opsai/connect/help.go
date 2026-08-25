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

package connect

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints ai connect --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro("Open a remote Ops Agent session (experimental). Does not use aissh interactive meta (!! / !ait ai).", ""); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AICLI("connect [ssh-options] destination")); err != nil {
		return err
	}
	return b.WriteExamples(
		toolname.AICLI("connect user@host"),
		toolname.AICLI("connect -J jump user@final"),
		toolname.AICLI("user@host"),
	)
}
