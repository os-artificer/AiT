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

package aimode

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func unknownModeErr(mode string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown mode %q", mode),
		"Switch AI interaction mode.",
		"mode ask",
		"mode agent",
		"help",
	)
}

func unknownApproveErr(approve string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown approve %q", approve),
		"Set how often commands are confirmed.",
		"approve each",
		"approve session",
		"trust",
		"help",
	)
}

func unknownPolicyErr(policy string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown policy %q", policy),
		"Set exec confirmation policy for the AI session.",
		"policy per-cmd",
		"policy once",
		"help",
	)
}

func usageRunCommandErr() error {
	return errhint.WithTryMsg("usage: run <command>",
		"Confirm and run a shell command from AI mode.",
		"run uptime",
		"help",
	)
}

func usageAitSubcommandErr() error {
	return errhint.WithTryMsg("usage: ait <subcommand> [args...]",
		"Run a local ait subcommand from AI mode.",
		"ait list",
		"ait session --list",
		"help",
	)
}

func aitSubcmdLocalOnlyErr() error {
	return errhint.WithTryMsg("ait subcommands are local-only",
		"Use run for remote shell commands while connected over SSH.",
		"run <command>",
		"help",
	)
}

func remoteRunnerNotConfiguredErr() error {
	return errhint.WithTryMsg("remote runner not configured",
		"Remote AI mode requires an active SSH session.",
		"help",
	)
}

func localRunnerNotConfiguredErr() error {
	return errhint.WithTryMsg("local runner not configured",
		"Local command execution is unavailable in this AI session.",
		"help",
	)
}

func cliExecutorNotConfiguredErr() error {
	return errhint.WithTryMsg("CLI executor not configured",
		"Internal wiring error; restart ait.",
	)
}
