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

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func runHistory(_ context.Context, env Env, _ []string) (bool, error) {
	if env.ParseLine == nil {
		return false, nil
	}

	words, err := env.ParseLine(env.Line)
	if err != nil {
		return false, err
	}

	var args []string
	if len(words) > 1 {
		args = words[1:]
	}
	return false, terminal.RunHistoryCommand(env.Out, env.Hist, args)
}
