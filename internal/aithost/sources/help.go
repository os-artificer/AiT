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

package sources

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const sourcesFlags = "      --json               emit JSON array\n"

// WriteHelp prints aithost sources --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro("List host metrics collectors and platform notes."); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AIHost + " sources [flags]"); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, sourcesFlags); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AIHost+" sources",
		toolname.AIHost+" sources --json",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
