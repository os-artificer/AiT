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
	"bytes"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestIsNativePromptLike_ANSI (P4) matches oh-my-zsh prompts with color codes.
func TestIsNativePromptLike_ANSI(t *testing.T) {
	line := "\x1b[1mroot@ylgeeker\x1b[0m [~] \x1b[1;33m➜\x1b[0m "
	if !isNativePromptLike(line) {
		t.Fatalf("expected ANSI native prompt match, line=%q", line)
	}
}

// TestRemoveManagedNativePromptLines_MultiLine (P5) strips native prompt in command output chunks.
func TestRemoveManagedNativePromptLines_MultiLine(t *testing.T) {
	chunk := []byte("ls\r\n\x1b[1mroot@ylgeeker\x1b[0m [~] \x1b[1;33m➜\x1b[0m ")
	out, removed := removeManagedNativePromptLines(chunk)
	if !removed {
		t.Fatal("expected native prompt removal")
	}
	if !strings.Contains(string(out), "ls") {
		t.Fatalf("command output should remain, got %q", out)
	}
	if strings.Contains(string(out), "ylgeeker") {
		t.Fatalf("native prompt should be removed, got %q", out)
	}
}

// TestFilterRemoteOutputChunkRepaint_SuppressesNativeAfterSetup (P6) repaints AiT prompt locally.
func TestFilterRemoteOutputChunkRepaint_SuppressesNativeAfterSetup(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)

	native := "\x1b[1mroot@ylgeeker\x1b[0m [~] \x1b[1;33m➜\x1b[0m \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), "root", "ylgeeker.com")
	if out != nil {
		t.Fatalf("native-only chunk should be fully filtered, got %q", out)
	}
	if !repaint {
		t.Fatal("expected repaint flag after native strip")
	}

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.repaintInteractivePrompt("root", "ylgeeker.com")
	if !strings.Contains(buf.String(), "AiT root@ylgeeker.com:") {
		t.Fatalf("repaint should show AiT prompt, got %q", buf.String())
	}
}

// TestFilterRemoteOutputChunkRepaint_StripsNativeWhenPromptShownOnly (P10) strips native
// lines once remotePromptShown even without setup sentinel (client-painted prompt).
func TestFilterRemoteOutputChunkRepaint_StripsNativeWhenPromptShownOnly(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)

	native := "root@ylgeeker [~] ➜ \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), "root", "ylgeeker.com")
	if out != nil {
		t.Fatalf("native chunk should be filtered when prompt shown, got %q", out)
	}
	if !repaint {
		t.Fatal("expected repaint after native strip")
	}
}

// TestFilterRemoteOutputChunkRepaint_KeepsNativeWithoutSetup (P7) does not strip native lines before setup completes.
func TestFilterRemoteOutputChunkRepaint_KeepsNativeWithoutSetup(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")

	native := "root@ylgeeker [~] ➜ \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), "root", "ylgeeker.com")
	if repaint {
		t.Fatal("should not repaint without managed setup")
	}
	if out == nil || !strings.Contains(string(out), "ylgeeker") {
		t.Fatalf("native prompt should pass through before setup, got %v", out)
	}
}

// TestStripUnwantedPromptLines_KeepsTableDropsPrompt (P9) keeps session list rows when a
// multi-line chunk also contains a native prompt echo line.
func TestStripUnwantedPromptLines_KeepsTableDropsPrompt(t *testing.T) {
	table := "#  ID        HOST           USER  STATE\r\n1  abc12345  ylgeeker.com   root  active\r\n"
	native := "\x1b[1mroot@ylgeeker\x1b[0m [~] \x1b[1;33m➜\x1b[0m \r\n"
	chunk := []byte(table + native)

	out, nativeRemoved := stripUnwantedPromptLines(chunk, "root", "ylgeeker.com", true, true)
	if !nativeRemoved {
		t.Fatal("expected prompt line removal")
	}
	if !strings.Contains(string(out), "abc12345") {
		t.Fatalf("session table should remain, got %q", out)
	}
	if strings.Contains(string(out), "ylgeeker") && !strings.Contains(string(out), "ylgeeker.com") {
		t.Fatalf("native prompt host should be removed, got %q", out)
	}
}

// TestFilterRemoteOutputChunkRepaint_RepaintsAfterAITDuplicateStrip (P11) repaints when
// a remote AiT prompt echo is dropped as duplicate after command output.
func TestFilterRemoteOutputChunkRepaint_RepaintsAfterAITDuplicateStrip(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)

	aitLine := prompt.RemoteFormat("root", "ylgeeker.com", "~", "")
	chunk := []byte("go  ips  snap\r\n" + aitLine)
	out, repaint := c.filterRemoteOutputChunkRepaint(chunk, "root", "ylgeeker.com")
	if out == nil || !strings.Contains(string(out), "go") {
		t.Fatalf("command output should remain, got %q", out)
	}
	if strings.Contains(string(out), "AiT root@ylgeeker.com:") {
		t.Fatalf("duplicate AiT prompt should be stripped, got %q", out)
	}
	if !repaint {
		t.Fatal("expected repaint after duplicate AiT prompt strip")
	}

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.repaintInteractivePrompt("root", "ylgeeker.com")
	if !strings.Contains(buf.String(), "AiT root@ylgeeker.com:") {
		t.Fatalf("repaint should show AiT prompt, got %q", buf.String())
	}
}

// TestIsDuplicatePromptLine_AIT (P8) still recognizes AiT prompt lines with ANSI stripped.
func TestIsDuplicatePromptLine_AIT(t *testing.T) {
	line := prompt.RemoteFormat("root", "ylgeeker.com", "~", "")
	if !isDuplicatePromptLine(line, "root", "ylgeeker.com") {
		t.Fatalf("AiT prompt line should be duplicate: %q", line)
	}
}
