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

package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ssh/config/edit"
	cfgget "github.com/os-artificer/AiT/internal/ssh/config/get"
	cfginit "github.com/os-artificer/AiT/internal/ssh/config/init"
	"github.com/os-artificer/AiT/internal/ssh/config/set"
	"github.com/os-artificer/AiT/internal/ssh/config/validate"
	"github.com/os-artificer/AiT/internal/ssh/config/view"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

func TestWriteConfigHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteConfigHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:",
		pkghelpdoc.SectionSubcommands,
		pkghelpdoc.SectionKeys,
		"Examples:",
		"help for config",
		"init",
		"view",
		"edit",
		"validate",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("config help missing %q:\n%s", want, out)
		}
	}
}

func TestWriteConfigChildHelp(t *testing.T) {
	writers := []struct {
		name string
		fn   func(*bytes.Buffer) error
	}{
		{"set", func(b *bytes.Buffer) error { return set.WriteHelp(b) }},
		{"get", func(b *bytes.Buffer) error { return cfgget.WriteHelp(b) }},
		{"init", func(b *bytes.Buffer) error { return cfginit.WriteHelp(b) }},
		{"view", func(b *bytes.Buffer) error { return view.WriteHelp(b) }},
		{"edit", func(b *bytes.Buffer) error { return edit.WriteHelp(b) }},
		{"validate", func(b *bytes.Buffer) error { return validate.WriteHelp(b) }},
	}
	for _, w := range writers {
		var buf bytes.Buffer
		if err := w.fn(&buf); err != nil {
			t.Fatalf("%s: %v", w.name, err)
		}
		out := buf.String()
		if !strings.Contains(out, "Usage:") {
			t.Fatalf("%s help missing Usage: %q", w.name, out)
		}
		if !strings.Contains(out, "help for "+w.name) {
			t.Fatalf("%s help missing footer: %q", w.name, out)
		}
	}
}
