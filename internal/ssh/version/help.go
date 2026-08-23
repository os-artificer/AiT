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

package version

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const versionIntroMain = "Print build version information (version, git commit hash, and build date).\n" +
	"Version is resolved from ldflags, the embedded VERSION file, or \"dev\" as fallback."

var versionUsage = toolname.CLI("version")

// WriteVersionHelp prints aissh version --help to w.
func WriteVersionHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(versionIntroMain, ""); err != nil {
		return err
	}
	if err := b.WriteUsage(versionUsage); err != nil {
		return err
	}
	return b.Finish("version")
}
