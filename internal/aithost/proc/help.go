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

package proc

import (
	"io"

	"github.com/os-artificer/AiT/internal/aithost/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aithost proc --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro(
		"List top processes sorted by CPU, then MEM, NET, and DISK rates.",
		"Default --limit is 10. Per-process network rates may be 0 on some platforms.",
	); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AIHost + " proc [flags]"); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, shared.SectionFlagsProc); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AIHost+" proc",
		toolname.AIHost+" proc --limit 5 --sort mem",
		toolname.AIHost+" proc --json",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
