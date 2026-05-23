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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// ErrChatMessageRequired is returned when a chat command has no message text.
var ErrChatMessageRequired = fmt.Errorf("AiT: chat requires a message (usage: ait ai chat <message> or !ait ai chat <message>)")

// parseChatCommand reports whether line is a one-shot chat command (chat <message>).
// Returns the message tail when ok is true.
func parseChatCommand(line string) (message string, ok bool) {
	word, rest := firstWord(normalizeLine(line))
	if word != replCmdChat {
		return "", false
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", false
	}
	return rest, true
}

// isChatCommandWithoutMessage reports whether line is exactly "chat" with no message.
func isChatCommandWithoutMessage(line string) bool {
	word, rest := firstWord(normalizeLine(line))
	return word == replCmdChat && strings.TrimSpace(rest) == ""
}

// RunChatOnce sends one user message and writes the assistant reply to opt.Out.
// Does not enter the REPL. Returns ErrAINotConfigured when opt.AI is nil, or ErrChatMessageRequired when message is empty.
func RunChatOnce(ctx context.Context, opt Options, message string) error {
	if opt.AI == nil {
		return ErrAINotConfigured
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return ErrChatMessageRequired
	}

	opt = fillOptionDefaults(opt)

	env := ai.AssistEnv{Scope: opt.Scope, HostLabel: opt.HostLabel, RemoteCwd: opt.RemoteCwd}
	if opt.Scrollback != nil {
		env.Scrollback = opt.Scrollback()
	}

	reply, _, err := opt.AI.Chat(ctx, nil, message, env)
	if err != nil {
		reportAIError(opt.ErrOut, err)
		return err
	}

	writeChatReply(opt, reply)
	return nil
}

func writeChatReply(opt Options, reply string) {
	if opt.Out == nil {
		return
	}
	if opt.Scope == ai.ScopeRemote {
		aimWrite(opt.Out, reply+"\r\n")
		return
	}
	_, _ = io.WriteString(opt.Out, reply+"\n")
}
