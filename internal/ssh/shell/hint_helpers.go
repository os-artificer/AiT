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

package shell

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func unclosedQuoteErr() error {
	return errhint.WithTryMsg("unclosed quote",
		"Close single-quoted segments in shell commands.",
		"help",
	)
}

func commandNotFoundErr(name string) error {
	binDir, err := platform.BinDir()
	if err != nil {
		binDir = "<AppRoot>/bin"
	}
	cause := fmt.Errorf("%w: %s", ErrCommandNotFound, name)
	return errhint.WithTry(cause,
		"Drop an executable in "+binDir+"/"+name+", use a "+toolname.AISsh+" subcommand, or ensure it is on PATH.",
		"help",
		"ls "+binDir,
	)
}
