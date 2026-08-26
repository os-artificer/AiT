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

package ait

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/version"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

type manageHelpSpec struct {
	usage string
	desc  string
	lines []string
}

var manageHelpTopics = map[string]manageHelpSpec{
	aitplugin.CmdInstall: {
		usage: "install [plugin...]",
		desc:  "Install official plugins (aissh, ailog, aiimg, aihost) from the latest GitHub Release.",
		lines: []string{
			"install",
			"install aissh",
			"install aissh ailog",
		},
	},
	aitplugin.CmdUpdate: {
		usage: "update [plugin...]",
		desc:  "Update installed official plugins from the latest GitHub Release.",
		lines: []string{
			"update",
			"update aissh",
		},
	},
	aitplugin.CmdUpgrade: {
		usage: "upgrade",
		desc:  "Update ait and all installed official plugins from the latest GitHub Release.",
		lines: []string{
			"upgrade",
		},
	},
}

func runManage(ctx context.Context, action aitrelease.Action, names []string, out io.Writer) error {
	opts := aitrelease.Options{
		UserAgent: fmt.Sprintf("AiT-ait/%s (%s)", version.Short(), aitrelease.DefaultRepoURL()),
	}
	if base := strings.TrimSpace(os.Getenv(aitrelease.TestReleaseBaseURLEnv)); base != "" {
		opts.BaseURL = base
	}
	req := aitrelease.Request{Action: action, Names: names}
	if err := aitrelease.Execute(ctx, opts, req, out); err != nil {
		return wrapManageErr(err)
	}
	return nil
}

func wrapManageErr(err error) error {
	switch {
	case errors.Is(err, aitrelease.ErrNoRelease):
		return errhint.WithTry(err,
			"No GitHub release is published yet for this repository.",
			"make install",
			toolname.AIT+" (then) install",
		)
	case errors.Is(err, aitrelease.ErrNoAsset):
		return errhint.WithTry(err,
			"No release archive matches this OS and CPU architecture.",
			toolname.AIT+" version",
		)
	case errors.Is(err, aitrelease.ErrInstallRoot):
		return errhint.WithTry(err,
			"Cannot write to the directory containing the ait binary.",
			"make install",
			"go install "+toolname.AIT+"@latest",
		)
	case errors.Is(err, aitrelease.ErrNothingToUpdate):
		return errhint.WithTry(err,
			"No installed official plugins are available to update.",
			toolname.AIT+" (then) install",
			toolname.AIT+" (then) install aissh",
		)
	case errors.Is(err, aitrelease.ErrUnknownPlugin):
		return errhint.WithTry(err,
			"Only official plugins can be installed or updated from GitHub Releases.",
			toolname.AIT+" (then) help",
		)
	case errors.Is(err, aitrelease.ErrUpgradeExtraArgs):
		return errhint.WithTry(err,
			"upgrade updates ait and all installed official plugins; it accepts no plugin names.",
			toolname.AIT+" (then) upgrade",
		)
	case errors.Is(err, aitrelease.ErrBundleInvalid):
		return errhint.WithTry(err,
			"The release archive could not be applied.",
			toolname.AIT+" (then) update",
		)
	default:
		return err
	}
}
