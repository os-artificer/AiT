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

package extract

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/aitimg/version"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// newCmd builds the private Cobra command for the default extract route.
func newCmd() *cobra.Command {
	var (
		jsonOut   bool
		noGeocode bool
		allOut    bool
		showVer   bool
	)

	cmd := &cobra.Command{
		Use:           toolname.AIImg + " [flags] <file...>",
		Short:         "Extract image metadata (EXIF, GPS, device)",
		Long:          fmt.Sprintf("Run '%s --help' for usage and examples.", toolname.AIImg),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if showVer {
				_, err := fmt.Fprintln(c.OutOrStdout(), version.Version())
				return err
			}
			return runFiles(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), args, runOptions{
				JSON:      jsonOut,
				NoGeocode: noGeocode,
				All:       allOut,
			})
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON array to stdout")
	cmd.Flags().BoolVar(&noGeocode, "no-geocode", false, "skip reverse geocode; Location shows 0")
	cmd.Flags().BoolVar(&allOut, "all", false, "print full structured EXIF after summary")
	cmd.Flags().BoolVarP(&showVer, "version", "V", false, "print build version and exit")

	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
