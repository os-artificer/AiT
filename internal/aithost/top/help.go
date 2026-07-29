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

package top

import (
	"io"

	"github.com/os-artificer/AiT/internal/aithost/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aithost top --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro(
		"Open a live dashboard with CPU, MEM, DISK, NET, and PROC panels.",
		"Requires a TTY. Without one, use snapshot or proc instead.",
	); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AIHost + " top [flags]"); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, shared.SectionFlagsTop); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AIHost+" top",
		toolname.AIHost+" top --interval 2s --limit 10",
		toolname.AIHost+" top --sort mem",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
