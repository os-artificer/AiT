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
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

func TestAIMode_LLNotBuiltin(t *testing.T) {
	var buf bytes.Buffer
	client := testAIClient(t, "use aitssh shell for directory listings")
	opt := Options{
		Scope:  ai.ScopeLocal,
		AI:     client,
		Out:    &buf,
		ErrOut: io.Discard,
	}
	st := newSessionState(opt)
	st.interaction = ModeAsk

	exit, err := ExecuteLine(context.Background(), opt, st, "ll", nil)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	out := buf.String()
	if strings.HasPrefix(strings.TrimSpace(out), "total ") {
		t.Fatalf("ai REPL must not run shell ll builtin: %q", out)
	}
	if strings.Contains(out, " ..\n") || strings.Contains(out, " ..\r\n") {
		t.Fatalf("ai REPL output looks like dirlist: %q", out)
	}
}
