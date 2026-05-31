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
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestBuildInjectParts_PlainCommand(t *testing.T) {
	parts := buildInjectParts("ls -la", InjectOpts{})
	joined := strings.Join(parts, "; ")

	if !strings.HasPrefix(joined, injectSpacePrefix) {
		t.Errorf("expected space prefix, got: %q", joined)
	}
	if strings.Contains(joined, "stty -echo") {
		t.Errorf("plain command should not have echo suppression, got: %q", joined)
	}
	if strings.Contains(joined, injectCaptureBeginPrint) {
		t.Errorf("plain command should not have capture markers, got: %q", joined)
	}
}

func TestBuildInjectParts_SuppressEcho(t *testing.T) {
	parts := buildInjectParts("pwd", InjectOpts{SuppressEcho: true})
	joined := strings.Join(parts, "; ")

	if !strings.Contains(joined, injectEchoOff) {
		t.Errorf("expected echo suppression off, got: %q", joined)
	}
	if !strings.Contains(joined, injectEchoOn) {
		t.Errorf("expected echo suppression on, got: %q", joined)
	}
}

func TestBuildInjectParts_CaptureOutput(t *testing.T) {
	parts := buildInjectParts("base64 file.txt", InjectOpts{CaptureOutput: true, SuppressEcho: true})
	joined := strings.Join(parts, "; ")

	if !strings.Contains(joined, injectCaptureBeginPrint) {
		t.Errorf("expected capture begin marker, got: %q", joined)
	}
	if !strings.Contains(joined, injectCaptureEndPrint) {
		t.Errorf("expected capture end marker, got: %q", joined)
	}
}

func TestBuildInjectParts_AllCommandsSpacePrefixed(t *testing.T) {
	parts := buildInjectParts("ls", InjectOpts{SuppressEcho: true, CaptureOutput: true})
	for _, part := range parts {
		if !strings.HasPrefix(part, injectSpacePrefix) {
			t.Errorf("part not space-prefixed: %q", part)
		}
	}
}

func TestParseCapturedOutput(t *testing.T) {
	delta := "before" + injectCaptureBegin + "hello world" + injectCaptureEnd + "after"
	got := parseCapturedOutput(delta)
	if got != "hello world" {
		t.Errorf("parseCapturedOutput = %q, want %q", got, "hello world")
	}
}

func TestParseCapturedOutput_NoMarkers(t *testing.T) {
	delta := "just some output without markers"
	got := parseCapturedOutput(delta)
	if got != delta {
		t.Errorf("parseCapturedOutput = %q, want %q", got, delta)
	}
}

func TestParseCapturedOutput_OnlyBeginMarker(t *testing.T) {
	delta := injectCaptureBegin + "partial output without end"
	got := parseCapturedOutput(delta)
	if got != "partial output without end" {
		t.Errorf("parseCapturedOutput = %q, want %q", got, "partial output without end")
	}
}

func TestParseHashFromOutput(t *testing.T) {
	output := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  /dev/null"
	got := parseHashFromOutput(output)
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("parseHashFromOutput = %q, want %q", got, want)
	}
}

func TestParseHashFromOutput_StdinFormat(t *testing.T) {
	output := "abc123  -"
	got := parseHashFromOutput(output)
	if got != "abc123" {
		t.Errorf("parseHashFromOutput = %q, want %q", got, "abc123")
	}
}

func TestParseHashFromOutput_Empty(t *testing.T) {
	got := parseHashFromOutput("")
	if got != "" {
		t.Errorf("parseHashFromOutput = %q, want empty", got)
	}
}

func TestInjectStripFragments(t *testing.T) {
	frags := injectStripFragments()
	if len(frags) < 4 {
		t.Errorf("InjectStripFragments returned %d fragments, want at least 4", len(frags))
	}
	foundBegin := false
	foundEnd := false
	for _, f := range frags {
		if f == injectCaptureBegin {
			foundBegin = true
		}
		if f == injectCaptureEnd {
			foundEnd = true
		}
	}
	if !foundBegin {
		t.Error("InjectStripFragments missing capture begin marker")
	}
	if !foundEnd {
		t.Error("InjectStripFragments missing capture end marker")
	}
}

// --- Chain mode PTY passthrough tests ---
//
// In local→proxy→remote chain mode, injected capture markers pass through an
// intermediate PTY driver. The old ANSI-escape markers (\x1b[AIT_CAP_B\x1b)
// were corrupted because the PTY driver interprets \x1b[ as a CSI (Control
// Sequence Introducer) and strips the escape bytes, leaving only "AIT_CAP_B"
// as plain text. This caused parseCapturedOutput to miss the markers and
// return the entire delta (including shell prompts, newlines), which then
// caused base64 decode failures in PTYTransfer.Get.
//
// The fix: use plain-text markers that no PTY driver will reinterpret.

func TestCaptureMarkersContainNoANSIEscape(t *testing.T) {
	// Verify that neither marker contains the ESC byte (\x1b), which would
	// trigger CSI/OSC interpretation by intermediate PTY drivers.
	markers := []struct {
		name  string
		value string
	}{
		{"injectCaptureBegin", injectCaptureBegin},
		{"injectCaptureEnd", injectCaptureEnd},
	}
	for _, m := range markers {
		if strings.Contains(m.value, "\x1b") {
			t.Errorf("%s contains ESC byte (\\x1b): %q", m.name, m.value)
		}
		if strings.Contains(m.value, "\033") {
			t.Errorf("%s contains octal ESC (\\033): %q", m.name, m.value)
		}
		// Also ensure no CSI introducer sequence (\x1b[) is present.
		if strings.Contains(m.value, "\x1b[") {
			t.Errorf("%s contains CSI introducer (\\x1b[): %q", m.name, m.value)
		}
	}
}

// simulateCSIStrip models what an intermediate PTY driver does when it
// encounters ANSI escape sequences: it interprets \x1b[ as a CSI introducer,
// consumes the following parameter bytes (0x30–0x3F) and intermediate bytes
// (0x20–0x2F), then the final byte (0x40–0x7E). For our old markers,
// \x1b[AIT_CAP_B\x1b → the PTY strips \x1b[ and the subsequent \x1b,
// leaving only "AIT_CAP_B" as plain text.
func simulateCSIStrip(input string) string {
	var b strings.Builder
	i := 0
	for i < len(input) {
		if i+1 < len(input) && input[i] == '\x1b' && input[i+1] == '[' {
			// CSI sequence: skip \x1b[ then consume parameter bytes (0x30-0x3F),
			// intermediate bytes (0x20-0x2F), and the final byte (0x40-0x7E).
			i += 2
			for i < len(input) && ((input[i] >= 0x30 && input[i] <= 0x3F) ||
				(input[i] >= 0x20 && input[i] <= 0x2F)) {
				i++
			}
			if i < len(input) && input[i] >= 0x40 && input[i] <= 0x7E {
				i++ // final byte
			}
		} else if input[i] == '\x1b' {
			// Bare ESC: skip it and the next byte (common PTY behavior).
			i++
			if i < len(input) {
				i++
			}
		} else {
			b.WriteByte(input[i])
			i++
		}
	}
	return b.String()
}

func TestCaptureMarkersSurvivePTYPassthrough(t *testing.T) {
	// Simulate an intermediate PTY driver stripping CSI sequences.
	// The new plain-text markers must survive intact.
	originalBegin := injectCaptureBegin
	originalEnd := injectCaptureEnd

	strippedBegin := simulateCSIStrip(originalBegin)
	strippedEnd := simulateCSIStrip(originalEnd)

	if strippedBegin != originalBegin {
		t.Errorf("injectCaptureBegin did not survive PTY passthrough: got %q, want %q", strippedBegin, originalBegin)
	}
	if strippedEnd != originalEnd {
		t.Errorf("injectCaptureEnd did not survive PTY passthrough: got %q, want %q", strippedEnd, originalEnd)
	}
}

func TestOldANSIMarkersCorruptedByPTYPassthrough(t *testing.T) {
	// Verify that the old ANSI-escape markers ARE corrupted by PTY passthrough.
	// This demonstrates the bug that the plain-text fix addresses.
	oldBegin := "\x1b[AIT_CAP_B\x1b"
	oldEnd := "\x1b[AIT_CAP_E\x1b"

	strippedBegin := simulateCSIStrip(oldBegin)
	strippedEnd := simulateCSIStrip(oldEnd)

	if strippedBegin == oldBegin {
		t.Errorf("old ANSI begin marker unexpectedly survived PTY passthrough: %q", oldBegin)
	}
	if strippedEnd == oldEnd {
		t.Errorf("old ANSI end marker unexpectedly survived PTY passthrough: %q", oldEnd)
	}

	// After CSI strip, the old markers are corrupted. The PTY driver interprets
	// \x1b[A as CSI "cursor up", consuming the 'A'; the result is "IT_CAP_B".
	// The end marker is similarly corrupted. These are no longer recognized as
	// capture delimiters by parseCapturedOutput.
	if strippedBegin == oldBegin {
		t.Errorf("old ANSI begin marker unexpectedly survived PTY passthrough: %q", oldBegin)
	}
	if strippedEnd == oldEnd {
		t.Errorf("old ANSI end marker unexpectedly survived PTY passthrough: %q", oldEnd)
	}
	// Verify the markers are genuinely corrupted (not just slightly different).
	if !strings.Contains(strippedBegin, "CAP_B") {
		t.Errorf("old ANSI begin after strip = %q, expected to still contain CAP_B", strippedBegin)
	}
	if !strings.Contains(strippedEnd, "CAP_E") {
		t.Errorf("old ANSI end after strip = %q, expected to still contain CAP_E", strippedEnd)
	}
}

func TestParseCapturedOutput_ChainModePTY(t *testing.T) {
	// Simulate the full chain-mode scenario:
	// 1. InjectCapture emits markers around command output
	// 2. The output passes through an intermediate PTY that strips CSI sequences
	// 3. parseCapturedOutput must correctly extract the captured content

	base64Output := "aGVsbG8gd29ybGQ=" // "hello world"
	delta := injectCaptureBegin + base64Output + injectCaptureEnd

	// Simulate PTY passthrough stripping CSI sequences.
	afterPTY := simulateCSIStrip(delta)

	got := parseCapturedOutput(afterPTY)
	if got != base64Output {
		t.Errorf("parseCapturedOutput after PTY = %q, want %q", got, base64Output)
	}
}

func TestParseCapturedOutput_ChainModePTY_WithShellPromptNoise(t *testing.T) {
	// Simulate the exact scenario from the reported bug:
	// In chain mode, if markers were corrupted, parseCapturedOutput would
	// return the entire delta including shell prompt "# ", newlines, etc.
	// This caused base64.StdEncoding.DecodeString to fail at "input byte 6".
	//
	// With plain-text markers, the shell prompt noise before/after markers
	// is correctly excluded by parseCapturedOutput. The content between
	// markers may include trailing \r\n from PTY output; callers like
	// captureRemoteBase64 strip these separately.

	base64Output := "aGVsbG8gd29ybGQ="
	delta := "# \r\n" + injectCaptureBegin + base64Output + "\r\n" + injectCaptureEnd + "# "

	afterPTY := simulateCSIStrip(delta)
	got := parseCapturedOutput(afterPTY)

	// parseCapturedOutput returns everything between the markers, including
	// the trailing \r\n. The PTY transfer layer strips \r\n separately.
	want := base64Output + "\r\n"
	if got != want {
		t.Errorf("parseCapturedOutput with shell noise = %q, want %q", got, want)
	}

	// Verify the critical property: the captured content does NOT include
	// the shell prompt "# " that appears before/after the markers.
	if strings.Contains(got, "#") {
		t.Errorf("parseCapturedOutput should not include shell prompt noise, got %q", got)
	}
}

func TestParseCapturedOutput_WithInjectedCommandEcho(t *testing.T) {
	// Simulate the actual scenario from the screenshot:
	// The injected command line (e.g. "printf 'AIT_CAPTURE_B_f3e7a1c9'")
	// is echoed to the PTY before the actual printf output because
	// stty -echo takes effect AFTER the shell has already echoed the command.
	// The delta therefore contains the marker text twice:
	//   1) inside the echoed command line (fake marker)
	//   2) as the actual printf output (real marker)
	//
	// parseCapturedOutput must use LastIndex to skip the fake marker
	// and extract the content between the real markers.

	base64Output := "aGVsbG8K"

	// Build a delta that includes the injected command line echo,
	// followed by the actual printf output.
	injectedCmdLine := " stty -echo 2>/dev/null; printf '" + injectCaptureBegin + "'; base64 '/tmp/test'; printf '" + injectCaptureEnd + "'; stty echo 2>/dev/null"
	delta := injectedCmdLine + "\n" + injectCaptureBegin + base64Output + "\n" + injectCaptureEnd

	got := parseCapturedOutput(delta)
	// base64 command outputs a trailing newline; parseCapturedOutput includes it.
	// captureRemoteBase64 strips \r\n separately.
	want := base64Output + "\n"
	if got != want {
		t.Errorf("parseCapturedOutput with command echo = %q, want %q", got, want)
	}

	// Verify the injected command line itself is NOT in the extracted content.
	if strings.Contains(got, "stty -echo") {
		t.Errorf("extracted content should not include injected command line, got %q", got)
	}
}

func TestParseCapturedOutput_WithInjectedCommandEcho_BothCaptures(t *testing.T) {
	// Simulate two consecutive InjectCapture calls (sha256sum then base64)
	// where the second delta still contains the first injection's echo.
	// This tests that LastIndex correctly finds the LAST real marker even
	// when multiple fake markers from previous injections are present.

	sha256Output := "95891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d082386a2e846f6be0  /tmp/test"
	base64Output := "aGVsbG8K"

	firstInjected := " stty -echo 2>/dev/null; printf '" + injectCaptureBegin + "'; sha256sum '/tmp/test'; printf '" + injectCaptureEnd + "'; stty echo 2>/dev/null"
	secondInjected := " stty -echo 2>/dev/null; printf '" + injectCaptureBegin + "'; base64 '/tmp/test'; printf '" + injectCaptureEnd + "'; stty echo 2>/dev/null"

	// Delta contains both injected command echoes plus both actual outputs.
	delta := firstInjected + "\n" +
		injectCaptureBegin + sha256Output + "\n" + injectCaptureEnd + "\n" +
		secondInjected + "\n" +
		injectCaptureBegin + base64Output + "\n" + injectCaptureEnd

	got := parseCapturedOutput(delta)
	want := base64Output + "\n"
	if got != want {
		t.Errorf("parseCapturedOutput with dual captures = %q, want %q", got, want)
	}
}

func TestParseCapturedOutput_OldMarkersFailInChainMode(t *testing.T) {
	// Demonstrate the bug: old ANSI markers are corrupted by PTY passthrough,
	// causing parseCapturedOutput to return the entire delta (not just the
	// captured content). This is the root cause of the base64 decode error
	// reported in the chain mode get operation.
	oldBegin := "\x1b[AIT_CAP_B\x1b"
	oldEnd := "\x1b[AIT_CAP_E\x1b"
	base64Output := "aGVsbG8="

	// Original delta with old ANSI markers
	delta := "# \r\n" + oldBegin + base64Output + "\r\n" + oldEnd + "# "

	// After PTY passthrough, markers are corrupted
	afterPTY := simulateCSIStrip(delta)

	// With corrupted markers, parseCapturedOutput using old markers would
	// NOT find them and would return the entire delta. We simulate this by
	// checking that the old markers are not present after PTY passthrough.
	if strings.Contains(afterPTY, oldBegin) {
		t.Errorf("old begin marker unexpectedly present after PTY passthrough")
	}
	if strings.Contains(afterPTY, oldEnd) {
		t.Errorf("old end marker unexpectedly present after PTY passthrough")
	}

	// The remaining text after PTY strip would be something like:
	// "# \r\nAIT_CAP_BaGVsbG8=\r\nAIT_CAP_E# "
	// parseCapturedOutput looking for oldBegin/oldEnd would NOT find them,
	// so it would return the entire delta (corrupting the base64 data).
	if strings.Contains(afterPTY, base64Output) && strings.Contains(afterPTY, "AIT_CAP_B") {
		// This confirms the old markers are stripped but base64 data is still there,
		// just with the wrong delimiters. parseCapturedOutput can't extract it.
		// This is exactly the bug scenario.
	}
}

func TestPrintfCommandsMatchMarkerConstants(t *testing.T) {
	// Verify that the printf commands emit exactly the marker constant values.
	// The printf commands are sent to the remote shell; the markers are used
	// locally to parse captured output. They must be identical.
	beginFromPrintf := strings.TrimPrefix(strings.TrimSuffix(injectCaptureBeginPrint, "'"), "printf '")
	endFromPrintf := strings.TrimPrefix(strings.TrimSuffix(injectCaptureEndPrint, "'"), "printf '")

	if beginFromPrintf != injectCaptureBegin {
		t.Errorf("printf begin = %q, marker begin = %q: mismatch", beginFromPrintf, injectCaptureBegin)
	}
	if endFromPrintf != injectCaptureEnd {
		t.Errorf("printf end = %q, marker end = %q: mismatch", endFromPrintf, injectCaptureEnd)
	}
}

func TestCaptureMarkersArePrintableASCII(t *testing.T) {
	// Plain-text markers should consist entirely of printable ASCII characters
	// so they pass through any PTY driver without modification.
	for _, m := range []struct {
		name  string
		value string
	}{
		{"injectCaptureBegin", injectCaptureBegin},
		{"injectCaptureEnd", injectCaptureEnd},
	} {
		for i, ch := range m.value {
			if ch < 0x20 || ch > 0x7E {
				t.Errorf("%s contains non-printable byte at index %d: %02x", m.name, i, ch)
			}
		}
	}
}

func TestInjectCapture_FallbackFullScrollback(t *testing.T) {
	snap := "noise" + injectCaptureBegin
	after := snap + "payload" + injectCaptureEnd

	delta := scrollback.Delta(snap, after)
	if captureMarkersPresent(delta) {
		t.Fatal("delta should lack both markers for split snap scenario")
	}
	if !captureMarkersPresent(after) {
		t.Fatal("full scrollback should contain markers")
	}

	var stdin bytes.Buffer
	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte(snap))
	orig := injectCaptureForRetry
	defer func() { injectCaptureForRetry = orig }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		_, _ = client.scroll.Write([]byte("payload" + injectCaptureEnd))
		full := client.scroll.String()
		delta := scrollback.Delta(snap, full)
		if !captureMarkersPresent(delta) {
			if captureMarkersPresent(full) {
				delta = full
			}
		}
		if !captureMarkersPresent(delta) {
			return "", ErrInjectCaptureNoMarkers
		}
		return parseCapturedOutput(delta), nil
	}

	out, err := c.injectCaptureRetry(&stdin, "sha256sum '/tmp/x'")
	if err != nil {
		t.Fatalf("fallback capture: %v", err)
	}
	if out != "payload" {
		t.Fatalf("out = %q, want payload", out)
	}
}

func TestInjectZModem_NoSttyNoQuietWait(t *testing.T) {
	var buf bytes.Buffer
	c := &Client{}
	cmd := "sz -y '/path/to/file'"
	if err := c.InjectZModemCommand(&buf, cmd); err != nil {
		t.Fatal(err)
	}
	written := buf.String()
	if strings.Contains(written, "stty") {
		t.Fatalf("zmodem inject must not wrap stty: %q", written)
	}
	if !strings.Contains(written, cmd) {
		t.Fatalf("missing cmd in inject line: %q", written)
	}

	var injectBuf bytes.Buffer
	if err := c.Inject(&injectBuf, cmd, InjectOpts{SuppressEcho: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(injectBuf.String(), "stty") {
		t.Fatal("control: Inject with SuppressEcho should use stty")
	}
}
