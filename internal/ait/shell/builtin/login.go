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

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
)

func runLogin(ctx context.Context, env Env, argv []string) (bool, error) {
	spec, err := openssh.ParseLogin(argv)
	if err != nil {
		return false, err
	}

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return false, err
	}

	return false, env.Mgr.Login(ctx, session.LoginOptsFromResolved(resolved))
}
