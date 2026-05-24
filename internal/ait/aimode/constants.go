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

// Commands — AI REPL built-in verbs (also checked in isMetaREPLCommand).
const (
	replCmdExit     = "exit"
	replCmdQuit     = "quit"
	replCmdHelp     = "help"
	replArgH        = "-h"
	replArgHelp     = "--help"
	replCmdClear    = "clear"
	replCmdHistory  = "history"
	replCmdPolicy   = "policy"
	replCmdModel    = "model"
	replCmdRun      = "run"
	replCmdAit      = "ait"
	replCmdChat     = "chat"
	replModelList   = "list"
	replModelStatus = "status"
)

// IO — CRLF output fragments and error prefix.
const (
	msgPrefixAiT               = "AiT: "
	msgSkippedCRLF             = "skipped\r\n"
	msgConversationClearedCRLF = "conversation cleared\r\n"
	errFmtUnknownPolicy        = "AiT: unknown policy %q (try per-cmd or once)"
	errUsageRunCommand         = "AiT: usage: run <command>"
	errRemoteRunnerNotConfig   = "AiT: remote runner not configured"
	errLocalRunnerNotConfig    = "AiT: local runner not configured"
	errUsageAitSubcommand      = "AiT: usage: ait <subcommand> [args...]"
	errCLIExecutorNotConfig    = "AiT: CLI executor not configured"
	errAitSubcmdLocalOnly      = "AiT: ait subcommands are local-only; use run for remote shell commands"
	errCannotSetOllamaProvider = "AiT: cannot set session provider to ollama (use tiered routing for local models)"
	msgExecPolicyPerCmdCRLF    = "exec policy: per_cmd (confirm each run)\r\n"
	msgExecPolicyOnceCRLF      = "exec policy: session_once (confirm once per AI session)\r\n"
	msgAIProviderSetFmt        = "AI provider set to %s\r\n"
)
