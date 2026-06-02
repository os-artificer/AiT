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

package client

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	openai "github.com/sashabaranov/go-openai"
)

// buildChatMessages constructs the full message list for a chat round: system prompt, history, and user message.
func buildChatMessages(scope Scope, history []ChatMessage, user string, env AssistEnv, meta PromptMeta) []openai.ChatCompletionMessage {
	out := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: systemPrompt(scope, env, meta)}}
	for _, m := range history {
		out = append(out, openai.ChatCompletionMessage{Role: m.Role, Content: m.Content})
	}

	userContent := user
	if len(history) == 0 {
		userContent = prependEnvContext(user, env)
	}
	out = append(out, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: userContent})
	return out
}

// assistMessages builds a minimal message pair for assist-style prompts.
func assistMessages(userPrompt, scrollback, remoteCwd string, meta PromptMeta) []openai.ChatCompletionMessage {
	env := AssistEnv{Scope: ScopeLocal, Scrollback: scrollback, RemoteCwd: remoteCwd}
	user := prependEnvContext(userPrompt, env)
	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt(ScopeLocal, env, meta)},
		{Role: openai.ChatMessageRoleUser, Content: user},
	}
}

// systemPrompt returns the system prompt for the given scope and mode.
func systemPrompt(scope Scope, env AssistEnv, meta PromptMeta) string {
	if env.AgentMode {
		return agentSystemPrompt(scope, env, meta)
	}
	base := fmt.Sprintf(`You are AiT, an AI assistant inside an SSH terminal client.
Suggest shell commands but do NOT claim you executed them.
Be concise. For security, never ask the user to paste secrets.

Routing facts (authoritative; do not guess):
ai_provider: %s
ai_model: %s
ai_mode: %s
When asked which provider or model you use, answer only from these fields.`,
		meta.Provider, meta.Model, meta.Mode)
	if scope == ScopeRemote {
		return base + `
The user is in remote AI mode (ask): suggested shell commands run on the remote host after they type "run <command>" and confirm.
Do not suggest running ait login from inside SSH; use run for remote shell work.`
	}
	return base + `
The user is in local AI mode (ask): use "run <command>" for local shell work or "ait <subcmd>" for AiT CLI after they confirm.`
}

// agentSystemPrompt returns the system prompt for Agent mode on the given scope.
func agentSystemPrompt(scope Scope, env AssistEnv, meta PromptMeta) string {
	scopeHint := "local machine"
	runTarget := "local shell"
	if scope == ScopeRemote {
		scopeHint = "remote SSH host"
		runTarget = "remote shell"
	}
	return fmt.Sprintf(`You are AiT in Agent mode on the %s.
Respond with a single JSON object only (no markdown), one of:
{"action":"run_shell","command":"<shell command>"}
{"action":"ait_cli","args":"<ait subcommand args without leading ait>"}
{"action":"answer","text":"<final reply to user>"}
Commands run on %s. Read-only commands (ls, cat, grep, ps, etc.) are auto-approved;
write/destructive commands require user confirmation. Do not run vim/less/top or other interactive TUI programs.
Never ask for secrets. Provider: %s, model: %s.`,
		scopeHint, runTarget, meta.Provider, meta.Model)
}

// prependEnvContext prepends terminal environment context (host, cwd, scrollback, OS) before the user message.
func prependEnvContext(user string, env AssistEnv) string {
	var parts []string
	if env.Scope == ScopeRemote {
		if env.HostLabel != "" {
			parts = append(parts, "remote_host: "+env.HostLabel)
		}
		if env.RemoteCwd != "" {
			parts = append(parts, "remote_cwd: "+env.RemoteCwd)
		}
	}
	parts = append(parts, "local_os: "+platform.OSDescription())
	if env.Scrollback != "" {
		parts = append(parts, "recent_terminal_output:\n"+truncate(env.Scrollback, 4000))
	}

	if len(parts) == 0 {
		return user
	}
	return strings.Join(parts, "\n\n") + "\n\n---\n\n" + user
}
