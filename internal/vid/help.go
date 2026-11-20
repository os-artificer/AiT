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

package aivid

import (
	"fmt"
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/vid/registry"
)

// WriteRootHelp prints aitvid --help to w.
func WriteRootHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro("Search video files or directories for text/objects; save hit keyframes."); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.VidCLI("<command> [flags]")); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, formatRootCommands(registry.RootHelpEntries())); err != nil {
		return err
	}
	examples := []string{
		toolname.VidCLI("search ./clip.mp4 --query plate --mode text --local-only"),
		toolname.VidCLI("search ./clips -r --query bike --camera-id cam03"),
		toolname.VidCLI("deps"),
		toolname.VidCLI("deps --install"),
		toolname.VidCLI("version"),
	}
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish(toolname.VidCLI(""))
}

func formatRootCommands(entries []pkghelpdoc.RootCommandEntry) string {
	var b strings.Builder
	for _, sc := range entries {
		fmt.Fprintf(&b, "  %-10s %s\n", sc.Name, sc.Short)
	}
	return b.String()
}
