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

package deps

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Name is the registered subcommand name.
const Name = "deps"

// New builds the `aitvid deps` command.
func New() *cobra.Command {
	var install, yes, jsonOut bool
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "Check or install ffmpeg/tesseract dependencies",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			need := deps.Need{Frames: true, OCR: true}
			rep := deps.Check(need, nil)
			if install {
				isTTY := term.IsTerminal(int(os.Stdin.Fd()))
				return deps.Install(context.Background(), deps.InstallOpts{
					Yes:     yes,
					IsTTY:   isTTY,
					NeedOCR: true,
					Confirm: func(prompt string) (bool, error) {
						fmt.Fprint(c.ErrOrStderr(), prompt)
						line, err := bufio.NewReader(os.Stdin).ReadString('\n')
						if err != nil {
							return false, err
						}
						s := strings.TrimSpace(strings.ToLower(line))
						return s == "y" || s == "yes", nil
					},
				})
			}
			if jsonOut {
				st := map[string]string{}
				for _, d := range rep.Deps {
					v := "missing"
					if d.OK {
						v = "ok"
					}
					st[string(d.Name)] = v
				}
				enc := json.NewEncoder(c.OutOrStdout())
				return enc.Encode(st)
			}
			_, err := fmt.Fprint(c.OutOrStdout(), deps.FormatStatus(rep))
			return err
		},
	}
	cmd.Flags().BoolVar(&install, "install", false, "install missing dependencies via package manager")
	cmd.Flags().BoolVar(&yes, "yes", false, "non-interactive install (required without a TTY)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON status object")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
