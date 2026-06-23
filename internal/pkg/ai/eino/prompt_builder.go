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

package eino

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/pkg/utf8tail"

	"github.com/cloudwego/eino/schema"
)

const (
	scrollbackContextMax = 4000

	promptKeyProvider = "ai_provider:"
	promptKeyModel    = "ai_model:"
	promptKeyMode     = "ai_mode:"

	envKeyRemoteHost = "remote_host:"
	envKeyRemoteCwd  = "remote_cwd:"
	envKeyLocalOS    = "local_os:"
	envKeyScrollback = "recent_terminal_output:\n"
	envContextSep    = "\n\n---\n\n"
)

// BuildChatMessages constructs Eino messages for a chat round.
func BuildChatMessages(scope apitypes.Scope, history []apitypes.ChatMessage, user string, env apitypes.AssistEnv, meta apitypes.PromptMeta) []*schema.Message {
	out := []*schema.Message{schema.SystemMessage(SystemPrompt(scope, env, meta))}
	for _, m := range history {
		if m.Role == RoleAssistant {
			out = append(out, schema.AssistantMessage(m.Content, nil))
			continue
		}
		out = append(out, schema.UserMessage(m.Content))
	}

	userContent := user
	if len(history) == 0 {
		userContent = PrependEnvContext(user, env)
	}
	out = append(out, schema.UserMessage(userContent))
	return out
}

// AssistMessages builds Eino messages for Assist-style prompts.
func AssistMessages(userPrompt, scrollback, remoteCwd string, meta apitypes.PromptMeta) []*schema.Message {
	env := apitypes.AssistEnv{Scope: apitypes.ScopeLocal, Scrollback: scrollback, RemoteCwd: remoteCwd}
	user := PrependEnvContext(userPrompt, env)
	return []*schema.Message{
		schema.SystemMessage(SystemPrompt(apitypes.ScopeLocal, env, meta)),
		schema.UserMessage(user),
	}
}

// SystemPrompt returns the system prompt for the given scope and mode.
func SystemPrompt(scope apitypes.Scope, env apitypes.AssistEnv, meta apitypes.PromptMeta) string {
	if env.AgentMode {
		return AgentSystemPrompt(scope, env, meta)
	}
	base := fmt.Sprintf(`You are AiT, an AI assistant inside an SSH terminal client.
Suggest shell commands but do NOT claim you executed them.
Be concise. For security, never ask the user to paste secrets.

Routing facts (authoritative; do not guess):
%s %s
%s %s
%s %s
When asked which provider or model you use, answer only from these fields.`,
		promptKeyProvider, meta.Provider, promptKeyModel, meta.Model, promptKeyMode, meta.Mode)
	if scope == apitypes.ScopeRemote {
		return base + `
The user is in remote AI mode (ask): suggested shell commands run on the remote host after they type "run <command>" and confirm.
Do not suggest running ` + toolname.CLI("login") + ` from inside SSH; use run for remote shell work.`
	}
	return base + `
The user is in local AI mode (ask): use "run <command>" for local shell work or "` + toolname.AISsh + ` <subcmd>" for AiT CLI after they confirm.`
}

// AgentSystemPrompt returns the system prompt for Agent mode on the given scope.
func AgentSystemPrompt(scope apitypes.Scope, env apitypes.AssistEnv, meta apitypes.PromptMeta) string {
	scopeHint := "local machine"
	runTarget := "local shell"
	if scope == apitypes.ScopeRemote {
		scopeHint = "remote SSH host"
		runTarget = "remote shell"
	}
	return fmt.Sprintf(`You are AiT in Agent mode on the %s.
Respond with a single JSON object only (no markdown), one of:
{"%s":"%s","%s":"<shell command>"}
{"%s":"%s","%s":"<%s subcommand args without leading %s>"}
{"%s":"%s","%s":"<final reply to user>"}
Commands run on %s. Read-only commands (ls, cat, grep, ps, etc.) are auto-approved;
write/destructive commands require user confirmation. Do not run vim/less/top or other interactive TUI programs.
Never ask for secrets. Provider: %s, model: %s.`,
		scopeHint,
		agent.JSONFieldAction, agent.ToolRunShell, agent.JSONFieldCommand,
		agent.JSONFieldAction, agent.ToolAitCLI, agent.JSONFieldArgs, toolname.AISsh, toolname.AISsh,
		agent.JSONFieldAction, agent.ToolAnswer, agent.JSONFieldText,
		runTarget, meta.Provider, meta.Model)
}

// PrependEnvContext prepends terminal environment context before the user message.
func PrependEnvContext(user string, env apitypes.AssistEnv) string {
	var parts []string
	if env.Scope == apitypes.ScopeRemote {
		if env.HostLabel != "" {
			parts = append(parts, envKeyRemoteHost+" "+env.HostLabel)
		}
		if env.RemoteCwd != "" {
			parts = append(parts, envKeyRemoteCwd+" "+env.RemoteCwd)
		}
	}
	parts = append(parts, envKeyLocalOS+" "+platform.OSDescription())
	if env.Scrollback != "" {
		parts = append(parts, envKeyScrollback+utf8tail.Truncate(env.Scrollback, scrollbackContextMax))
	}

	if len(parts) == 0 {
		return user
	}
	return strings.Join(parts, "\n\n") + envContextSep + user
}
