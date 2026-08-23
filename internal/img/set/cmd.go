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

package set

import (
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const (
	// Name is the top-level subcommand name.
	Name = "set"
)

type setFlags struct {
	make         string
	model        string
	software     string
	lensMake     string
	lensModel    string
	serial       string
	datetime     string
	gps          string
	altitude     string
	output       string
	inplace      bool
	yes          bool
	outputSet    bool
	sawMake      bool
	sawModel     bool
	sawSoftware  bool
	sawLensMake  bool
	sawLensModel bool
	sawSerial    bool
	sawDatetime  bool
	sawGPS       bool
	sawAltitude  bool
}

// New builds the `aiimg set` cobra command.
func New() *cobra.Command {
	var f setFlags

	cmd := &cobra.Command{
		Use:           Name + " [flags] <file...>",
		Short:         "Set writable EXIF fields on JPEG images",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f.sawMake = c.Flags().Changed("make")
			f.sawModel = c.Flags().Changed("model")
			f.sawSoftware = c.Flags().Changed("software")
			f.sawLensMake = c.Flags().Changed("lens-make")
			f.sawLensModel = c.Flags().Changed("lens-model")
			f.sawSerial = c.Flags().Changed("serial")
			f.sawDatetime = c.Flags().Changed("datetime")
			f.sawGPS = c.Flags().Changed("gps")
			f.sawAltitude = c.Flags().Changed("altitude")
			f.outputSet = c.Flags().Changed("output")
			return runSet(c.OutOrStdout(), c.ErrOrStderr(), args, f)
		},
	}

	cmd.Flags().StringVar(&f.make, "make", "", "camera make")
	cmd.Flags().StringVar(&f.model, "model", "", "camera model")
	cmd.Flags().StringVar(&f.software, "software", "", "software tag")
	cmd.Flags().StringVar(&f.lensMake, "lens-make", "", "lens make")
	cmd.Flags().StringVar(&f.lensModel, "lens-model", "", "lens model")
	cmd.Flags().StringVar(&f.serial, "serial", "", "body serial number")
	cmd.Flags().StringVar(&f.datetime, "datetime", "", "capture time (RFC3339 or YYYY:MM:DD HH:MM:SS)")
	cmd.Flags().StringVar(&f.gps, "gps", "", "latitude,longitude (decimal degrees)")
	cmd.Flags().StringVar(&f.altitude, "altitude", "", "altitude in meters")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write result to this path (single input)")
	cmd.Flags().BoolVar(&f.inplace, "inplace", false, "overwrite original file(s)")
	cmd.Flags().BoolVarP(&f.yes, "yes", "y", false, "same as --inplace (yes, overwrite originals)")

	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
