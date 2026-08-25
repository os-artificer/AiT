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

// Package deps holds shared dependencies for the ai command subtree.
package deps

import (
	"context"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

// CLIExecutor invokes aissh (or other toolbox CLIs) with custom IO.
type CLIExecutor func(args []string, out, errOut io.Writer) error

// RunWithMgr opens config, builds an AI client, and runs fn under a cancellable context.
type RunWithMgr func(fn func(ctx context.Context, cfg *config.Config, aiClient *pkgai.Client) error) error

// Deps groups cross-package dependencies for ai subcommands.
type Deps struct {
	Exec       CLIExecutor
	RunWithMgr RunWithMgr
	// ApplyOpsDomain sets Ops domain + toolbox on local agentcore config.
	ApplyOpsDomain func(agentcore.Config) agentcore.Config
}
