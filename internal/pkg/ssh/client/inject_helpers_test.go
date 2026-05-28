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
	"strings"
	"testing"
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
