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

package helpdoc

import (
	"bytes"
	"testing"
)

func TestWriteProviderEnableHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteProviderEnableHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(),
		"Usage:", "Examples:", "help for enable", "ollama", "--yes", "--no-install",
	)
}

func TestWriteProviderDisableHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteProviderDisableHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "help for disable", "minimax")
}

func TestWriteAIAskHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAIAskHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "ait ai ask", "stdin", "help for ask")
}

func TestWriteOllamaSetupHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteOllamaSetupHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "Examples:", "help for setup", "[y/N]")
}

func TestWriteAIOllamaHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAIOllamaHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "ait ai ollama", "status", "start", "disable")
}

func TestWriteOllamaStatusHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteOllamaStatusHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "Usage:", "ait ai ollama status", "tiered.enabled")
}

func TestRootHelpListsAI(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	helpMustContain(t, buf.String(), "ai", "provider")
}
