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

package builtin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/dirlist"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func runLL(_ context.Context, env Env, argv []string) (bool, error) {
	if len(argv) > 1 {
		return false, llUsageErr()
	}

	dir, err := llTargetDir(argv)
	if err != nil {
		return false, err
	}

	var buf bytes.Buffer
	if err := dirlist.WriteLongListing(&buf, dir); err != nil {
		return false, llPathErr(dir, err)
	}
	writeTTY(env.Out, buf.String())
	return false, nil
}

func llTargetDir(argv []string) (string, error) {
	if len(argv) == 0 {
		return os.Getwd()
	}
	return platform.ExpandUserAbsPath(argv[0])
}

func llUsageErr() error {
	return errhint.WithTryMsg("usage: ll [path]",
		fmt.Sprintf("List the current directory or one path in ls -la style (%s only).", toolname.CLI("shell")),
		"ll",
		"ll ..",
		"help",
	)
}

func llNotFoundErr(path string) error {
	return errhint.WithTryMsg("ll: no such file or directory: "+path,
		"Check the path or list the parent directory.",
		"ll",
		"ll ..",
		"help",
	)
}

func llPathErr(dir string, err error) error {
	if errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return llNotFoundErr(dir)
	}
	return err
}
