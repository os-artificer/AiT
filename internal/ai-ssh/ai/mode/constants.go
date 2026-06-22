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

package mode

import (
	"github.com/os-artificer/AiT/internal/ai-ssh/ai/mode/chat"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Commands — AI REPL built-in verbs (also checked in isMetaREPLCommand).
const (
	replCmdPolicy   = "policy"
	replCmdApprove  = "approve"
	replCmdTrust    = "trust"
	replCmdMode     = "mode"
	replCmdModel    = "model"
	replCmdRun      = "run"
	replCmdAISsh    = toolname.AISsh
	replCmdAsk      = chat.Verb
	replModelList   = "list"
	replModelStatus = "status"
)

// IO — CRLF output fragments and error prefix.
const (
	msgPrefixAiT               = errhint.PrefixAiT
	msgSkippedCRLF             = "skipped\r\n"
	msgConversationClearedCRLF = "conversation cleared\r\n"
	errRemoteRunnerNotConfig   = "AiT: remote runner not configured"
	errLocalRunnerNotConfig    = "AiT: local runner not configured"
	errCLIExecutorNotConfig    = "AiT: CLI executor not configured"
	errAitSubcmdLocalOnly      = "AiT: " + toolname.AISsh + " subcommands are local-only; use run for remote shell commands"
	errCannotSetOllamaProvider = "AiT: cannot set session provider to ollama (use tiered routing for local models)"
	msgExecPolicyPerCmdCRLF    = "exec policy: per_cmd (confirm each run)\r\n"
	msgExecPolicyOnceCRLF      = "exec policy: session_once (confirm once per AI session)\r\n"
	msgApproveEachCRLF         = "approve: each — confirm before every command\r\n"
	msgApproveSessionCRLF      = "approve: session — trust safe commands after first yes\r\n"
	msgModeAskCRLF             = "mode: ask — Q&A only; use run to execute commands\r\n"
	msgModeAgentCRLF           = "mode: agent — read commands auto-run; write commands need confirmation\r\n"
	msgPolicyDeprecated        = "AiT: policy is deprecated; use approve each or approve session"
	msgInteractiveSkipCRLF     = "interactive command not run in Agent mode; use shell or run manually\r\n"
	msgAIProviderSetFmt        = "AI provider set to %s\r\n"
	agentMaxSteps              = 16
	msgAgentStepLimitFmt       = "AiT: agent step limit reached (max %d steps). " +
		"Use 'mode ask' for complex tasks, or set ai.agent.max_steps in config.\r\n"
)
