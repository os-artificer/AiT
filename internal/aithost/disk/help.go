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

package disk

import (
	"io"

	"github.com/os-artificer/AiT/internal/aithost/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aithost disk --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro("Show host disk usage and IO rates."); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AIHost + " disk [flags]"); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, shared.SectionFlagsJSONWatch); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AIHost+" disk",
		toolname.AIHost+" disk --json",
		toolname.AIHost+" disk --watch --interval 2s",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
