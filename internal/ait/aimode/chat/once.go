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

// Package chat exposes the one-shot AI chat path that powers
// `ait ai chat <message>` and the `chat …` meta-REPL command.
package chat

import (
	"context"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// Verb is the REPL command word that triggers a one-shot chat.
const Verb = "chat"

// Input collects what RunOnce needs without depending on aimode.Options.
type Input struct {
	// AI is the chat client; required.
	AI *ai.Client
	// Message is the user turn; whitespace is trimmed before sending.
	Message string
	// Scope distinguishes local (LF) vs remote (CRLF) reply formatting.
	Scope ai.Scope
	// HostLabel / RemoteCwd populate ai.AssistEnv for the model.
	HostLabel string
	RemoteCwd string
	// Scrollback, when non-nil, contributes recent terminal output to the env.
	Scrollback func() string
	// Out receives the assistant reply; nil writers are skipped.
	Out io.Writer
}

// RunOnce sends one user message and writes the assistant reply.
// Does not enter a REPL or store history.
// Returns ai.ErrNotConfigured when in.AI is nil, ai.ErrChatMessageRequired
// when in.Message is blank, or the underlying ai.Chat error.
func RunOnce(ctx context.Context, in Input) error {
	if in.AI == nil {
		return ai.NotConfiguredErr()
	}

	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		return ai.ChatMessageRequiredErr()
	}

	env := ai.AssistEnv{
		Scope:     in.Scope,
		HostLabel: in.HostLabel,
		RemoteCwd: in.RemoteCwd,
	}
	if in.Scrollback != nil {
		env.Scrollback = in.Scrollback()
	}

	reply, _, _, err := in.AI.Chat(ctx, nil, msg, env)
	if err != nil {
		return err
	}

	writeReply(in.Out, in.Scope, reply)
	return nil
}

// ParseCommand reports whether line is a one-shot chat command ("chat <message>").
// Returns the trimmed message tail when ok is true.
// Returns (_, false) for bare "chat" with no message.
func ParseCommand(line string) (message string, ok bool) {
	rest, isVerb := splitVerb(line)
	if !isVerb {
		return "", false
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", false
	}
	return rest, true
}

// IsEmptyCommand reports whether line is exactly "chat" with no message argument.
func IsEmptyCommand(line string) bool {
	rest, isVerb := splitVerb(line)
	return isVerb && strings.TrimSpace(rest) == ""
}

func splitVerb(line string) (rest string, isVerb bool) {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}
	i := strings.IndexByte(line, ' ')
	if i < 0 {
		return "", line == Verb
	}
	if line[:i] != Verb {
		return "", false
	}
	return strings.TrimSpace(line[i+1:]), true
}

func writeReply(out io.Writer, scope ai.Scope, reply string) {
	if out == nil {
		return
	}
	if scope == ai.ScopeRemote {
		_, _ = io.WriteString(out, strings.ReplaceAll(reply, "\n", "\r\n")+"\r\n")
		return
	}
	_, _ = io.WriteString(out, reply+"\n")
}
