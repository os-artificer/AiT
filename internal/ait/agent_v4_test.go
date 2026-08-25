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

package ait_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// TestATV4_ReservedAgentName (AT-V4-1) agent is reserved; manifest name agent rejected.
func TestATV4_ReservedAgentName(t *testing.T) {
	names := aitplugin.ReservedCommandNames()
	found := false
	for _, n := range names {
		if n == aitplugin.CmdAgent {
			found = true
		}
	}
	if !found {
		t.Fatal("agent must be in reserved names")
	}
	if !aitplugin.IsReservedName("agent") {
		t.Fatal("IsReservedName(agent) should be true")
	}
}

// TestExecuteWriters_AgentHelp (AT-V4-3) top-level ait agent without TTY fails fast when AI gated off.
func TestExecuteWriters_AgentHelp(t *testing.T) {
	t.Setenv("AIT_AI_EXPERIMENTAL", "")
	config.SetAIExperimentalCLI(false)
	err := ait.ExecuteWriters([]string{"agent"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected experimental gate")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "experimental") {
		t.Fatalf("expected experimental error, got %v", err)
	}
}

// TestExecuteWriters_AgentExperimental (AG-8) experimental flag enables agent route (not experimental-blocked).
func TestExecuteWriters_AgentExperimental(t *testing.T) {
	t.Setenv("AIT_AI_EXPERIMENTAL", "1")
	config.SetAIExperimentalCLI(true)
	err := ait.ExecuteWriters([]string{"--ai-experimental", "agent"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "experimental") {
		t.Fatalf("should pass experimental gate, got %v", err)
	}
}

// TestRootHelp_IncludesAgent (AT-V4-2) help still lists agent reserved command.
func TestRootHelp_IncludesAgent(t *testing.T) {
	var buf bytes.Buffer
	if err := ait.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, aitplugin.CmdAgent) {
		t.Fatalf("help missing agent: %s", out)
	}
	if !strings.Contains(out, toolname.AIT) {
		t.Fatal("help missing ait usage")
	}
}
