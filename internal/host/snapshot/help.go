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

package snapshot

import (
	"io"

	"github.com/os-artificer/AiT/internal/host/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aihost snapshot --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro(
		"Print host identity, CPU, memory, disk, network, protocols, and top processes once.",
		"Use --json for machine-readable output (useful without a TTY).",
	); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AIHost + " snapshot [flags]"); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, shared.SectionFlagsProc); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AIHost+" snapshot",
		toolname.AIHost+" snapshot --json",
		toolname.AIHost+" snapshot --limit 5",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
