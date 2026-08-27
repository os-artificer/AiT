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

package agentcore

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/approve"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	pkgagent "github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/ai/client"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// RunOnce runs a single agent turn for question until the model answers or step limit is reached.
// Returns ai.ErrNotConfigured when cfg.AI is nil or ai.ErrChatMessageRequired when question is blank.
func RunOnce(ctx context.Context, cfg Config, question string) error {
	question = strings.TrimSpace(question)
	if question == "" {
		return ai.ChatMessageRequiredErr()
	}
	if cfg.AI == nil {
		return ai.NotConfiguredErr()
	}

	cfg = fillDefaults(cfg)
	st := newSessionState()
	emit(cfg, Event{Kind: EventSessionStarted})
	emit(cfg, Event{Kind: EventGoalReceived, Detail: question})
	err := runAgentTurn(ctx, cfg, st, question)
	if err == nil {
		emit(cfg, Event{Kind: EventSessionEnded})
	}
	return err
}

// RunLoop runs the agent task loop until exit/stop/back or ctx cancel.
// initialGoal, if non-empty, is processed as the first user turn.
func RunLoop(ctx context.Context, cfg Config, initialGoal string) error {
	if cfg.AI == nil {
		return ai.NotConfiguredErr()
	}

	cfg = fillDefaults(cfg)
	st := newSessionState()
	emit(cfg, Event{Kind: EventSessionStarted})

	if g := strings.TrimSpace(initialGoal); g != "" {
		emit(cfg, Event{Kind: EventGoalReceived, Detail: g})
		if err := runAgentTurn(ctx, cfg, st, g); err != nil {
			return err
		}
	}

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) && (cfg.In == os.Stdin) {
		return runLoopTTY(ctx, cfg, st, fd)
	}
	return runLoopPlain(ctx, cfg, st)
}

func fillDefaults(cfg Config) Config {
	if cfg.In == nil {
		cfg.In = os.Stdin
	}
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	if cfg.ErrOut == nil {
		cfg.ErrOut = os.Stderr
	}
	if cfg.Emitter == nil {
		cfg.Emitter = NopEmitter
	}
	if cfg.Local == nil && cfg.Scope == ai.ScopeLocal {
		cfg.Local = defaultLocalRunner{}
	}
	if cfg.Capture == nil && cfg.Scope == ai.ScopeLocal {
		cfg.Capture = runner.DefaultLocalCaptureRunner()
	}
	return cfg
}

func runAgentTurn(ctx context.Context, cfg Config, st *sessionState, question string) error {
	if cfg.AI == nil {
		return ai.NotConfiguredErr()
	}

	env := assistEnv(cfg, st)
	userMsg := question
	maxSteps := cfg.MaxAgentSteps()
	parseState := newAgentParseRetryState(st.history)

	for step := 0; step < maxSteps; step++ {
		reply, _, updated, chatErr := cfg.AI.ChatAgent(ctx, st.history, userMsg, env)
		if chatErr != nil {
			return chatErr
		}
		if reply != "" {
			if cfg.StreamDelta != nil {
				cfg.StreamDelta(reply)
			}
			emit(cfg, Event{Kind: EventTokenDelta, Detail: reply})
		}

		parseResult := pkgagent.ResolveAgentReply(reply)
		if !parseResult.OK() {
			userMsg, chatErr = handleAgentParseFailure(st, &parseState, userMsg, reply, parseResult)
			if chatErr != nil {
				return chatErr
			}
			continue
		}

		canonical, err := parseResult.Envelope.CanonicalJSON()
		if err != nil {
			st.history = parseState.turnStartHistory
			return formatMalformedReplyErr()
		}
		st.history = client.ReplaceLastAssistant(updated, canonical)

		act := envelopeToAction(parseResult.Envelope)
		toolResult, done, err := executeAgentAction(ctx, cfg, st, act)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		userMsg = pkgagent.ToolResultPrefix + toolResult
	}

	emit(cfg, Event{Kind: EventStepLimitReached, Detail: fmt.Sprintf("%d", maxSteps)})
	agentWrite(cfg.Out, fmt.Sprintf(msgAgentStepLimitFmt, maxSteps))
	return waitContinueOrExit(ctx, cfg, st)
}

func waitContinueOrExit(ctx context.Context, cfg Config, st *sessionState) error {
	sc := bufio.NewScanner(cfg.In)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, _ = io.WriteString(cfg.Out, "AiT Agent> continue? ")
		if !sc.Scan() {
			return sc.Err()
		}
		line := strings.TrimSpace(strings.ToLower(sc.Text()))
		switch line {
		case "continue", "y", "yes":
			return runAgentTurn(ctx, cfg, st, "continue the previous task")
		case metaExit, metaStop, metaBack, "n", "no", "quit":
			emit(cfg, Event{Kind: EventSessionEnded})
			return nil
		default:
			if line != "" {
				emit(cfg, Event{Kind: EventGoalReceived, Detail: line})
				return runAgentTurn(ctx, cfg, st, line)
			}
		}
	}
}

func runLoopPlain(ctx context.Context, cfg Config, st *sessionState) error {
	sc := bufio.NewScanner(cfg.In)
	hist := terminal.NewReplLineHistory(cfg.ErrOut)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(cfg.Out, "AiT Agent> ")
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			emit(cfg, Event{Kind: EventSessionEnded})
			return nil
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		exit, err := handleLoopMeta(ctx, cfg, st, line, hist)
		if exit {
			emit(cfg, Event{Kind: EventSessionEnded})
			return err
		}
		if err != nil {
			agentPrintErr(cfg.ErrOut, err.Error())
			continue
		}
		if err := runAgentTurn(ctx, cfg, st, line); err != nil {
			return err
		}
	}
}

func runLoopTTY(ctx context.Context, cfg Config, st *sessionState, fd int) error {
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return runLoopPlain(ctx, cfg, st)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	br := bufio.NewReader(cfg.In)
	hist := terminal.NewReplLineHistory(cfg.ErrOut)
	if cfg.ConfirmRead == nil {
		cfg = cfg.withTTYConfirm(br)
	}
	cfg.SuspendTTY = suspendRawTTY(fd, oldState)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		scope, host, _, _, _ := resolveDynamic(cfg)
		prompt := loopPrompt(scope, host)
		line, err := terminal.ReadLine(br, cfg.Out, terminal.LineEditOptions{
			Prompt:        prompt,
			History:       hist,
			RecordHistory: true,
		})
		if err != nil {
			if err == io.EOF {
				emit(cfg, Event{Kind: EventSessionEnded})
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		exit, err := handleLoopMeta(ctx, cfg, st, line, hist)
		if exit {
			emit(cfg, Event{Kind: EventSessionEnded})
			return err
		}
		if err != nil {
			agentPrintErr(cfg.ErrOut, err.Error())
			continue
		}
		if err := runAgentTurn(ctx, cfg, st, line); err != nil {
			return err
		}
	}
}

func loopPrompt(scope ai.Scope, host string) string {
	if scope == ai.ScopeRemote && host != "" {
		return fmt.Sprintf("AiT Agent(%s)> ", host)
	}
	return "AiT Agent> "
}

func handleLoopMeta(ctx context.Context, cfg Config, st *sessionState, line string, hist *terminal.LineHistory) (exit bool, err error) {
	_ = ctx
	word, rest := firstWord(line)
	switch word {
	case metaExit, metaStop:
		st.Trust.Clear()
		return true, nil
	case metaBack:
		if cfg.Nested {
			st.Trust.Clear()
			return true, nil
		}
		return false, nil
	case "clear":
		st.history = nil
		agentWrite(cfg.Out, msgConversationClear)
		return false, nil
	case "history":
		parts := strings.Fields(line)
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}
		return false, terminal.RunHistoryCommand(terminal.NewCRLFWriter(cfg.Out), hist, args)
	case metaModel:
		return false, showModelStatus(cfg, rest)
	default:
		return false, nil
	}
}

func showModelStatus(cfg Config, rest string) error {
	rest = strings.TrimSpace(rest)
	if rest == "list" {
		if cfg.AI == nil {
			return ai.NotConfiguredErr()
		}
		lines, err := cfg.AI.ListModels(context.Background())
		if err != nil {
			return err
		}
		if lines != nil {
			_, _ = fmt.Fprintln(cfg.Out, lines.Format())
		}
		return nil
	}
	if cfg.AI != nil {
		_, _ = fmt.Fprintln(cfg.Out, cfg.AI.StatusLine())
	}
	return nil
}

func firstWord(line string) (word, rest string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", ""
	}
	i := strings.IndexByte(line, ' ')
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimSpace(line[i+1:])
}

func (cfg Config) withTTYConfirm(br *bufio.Reader) Config {
	if cfg.ConfirmRead != nil {
		return cfg
	}
	out := cfg.Out
	if out == nil {
		out = os.Stdout
	}
	cfg.ConfirmRead = func(prompt string) (approve.Result, error) {
		agentWrite(out, prompt)
		line, err := terminal.ReadTTYLineNoHistory(br, out, "")
		if err != nil {
			return approve.Result{}, err
		}
		allowPattern := !strings.Contains(prompt, "HIGH-RISK")
		return approve.ParseChoice(line, allowPattern, allowPattern), nil
	}
	return cfg
}
