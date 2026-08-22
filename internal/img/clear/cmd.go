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

package clear

import (
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const (
	// Name is the top-level subcommand name.
	Name = "clear"
)

type clearFlags struct {
	make      bool
	model     bool
	software  bool
	lensMake  bool
	lensModel bool
	serial    bool
	datetime  bool
	gps       bool
	altitude  bool
	all       bool
	output    string
	inplace   bool
	yes       bool
	outputSet bool
}

// New builds the `aiimg clear` cobra command.
func New() *cobra.Command {
	var f clearFlags

	cmd := &cobra.Command{
		Use:           Name + " [flags] <file...>",
		Short:         "Clear writable EXIF fields on JPEG images",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f.outputSet = c.Flags().Changed("output")
			return runClear(c.OutOrStdout(), c.ErrOrStderr(), args, f)
		},
	}

	cmd.Flags().BoolVar(&f.make, "make", false, "clear camera make")
	cmd.Flags().BoolVar(&f.model, "model", false, "clear camera model")
	cmd.Flags().BoolVar(&f.software, "software", false, "clear software tag")
	cmd.Flags().BoolVar(&f.lensMake, "lens-make", false, "clear lens make")
	cmd.Flags().BoolVar(&f.lensModel, "lens-model", false, "clear lens model")
	cmd.Flags().BoolVar(&f.serial, "serial", false, "clear body serial number")
	cmd.Flags().BoolVar(&f.datetime, "datetime", false, "clear capture time")
	cmd.Flags().BoolVar(&f.gps, "gps", false, "clear GPS latitude/longitude")
	cmd.Flags().BoolVar(&f.altitude, "altitude", false, "clear altitude")
	cmd.Flags().BoolVar(&f.all, "all", false, "clear all writable fields")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write result to this path (single input)")
	cmd.Flags().BoolVar(&f.inplace, "inplace", false, "overwrite original file(s)")
	cmd.Flags().BoolVarP(&f.yes, "yes", "y", false, "same as --inplace (yes, overwrite originals)")

	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
