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
	"context"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func runLoad(ctx context.Context, env Env, argv []string) (bool, error) {
	var id, password string
	var tmux bool
	var err error

	if len(argv) == 0 {
		in := env.In
		if in == nil {
			in = os.Stdin
		}
		id, err = env.Mgr.PickSessionFrom(in, env.Out)
		if err != nil {
			return false, err
		}
		terminal.DrainStdin()
	} else {
		id, password, tmux, err = parseLoadArgs(argv)
		if err != nil {
			return false, err
		}
	}

	return false, env.Mgr.Load(ctx, id, tmux, password)
}
