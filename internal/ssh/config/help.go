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
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ssh/config/edit"
	cfgget "github.com/os-artificer/AiT/internal/ssh/config/get"
	cfginit "github.com/os-artificer/AiT/internal/ssh/config/init"
	"github.com/os-artificer/AiT/internal/ssh/config/set"
	"github.com/os-artificer/AiT/internal/ssh/config/shared"
	"github.com/os-artificer/AiT/internal/ssh/config/subregistry"
	"github.com/os-artificer/AiT/internal/ssh/config/validate"
	"github.com/os-artificer/AiT/internal/ssh/config/view"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const configIntro = "Read and write AiT YAML configuration (session TTL, AI providers, API keys)."

// WriteConfigHelp prints aissh config --help to w.
func WriteConfigHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(configIntro); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.CLI("config")); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, formatConfigSubcommands()); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionKeys, shared.KeysHelp); err != nil {
		return err
	}
	if err := b.WriteExamples(
		toolname.CLI("config init"),
		toolname.CLI("config view"),
		toolname.CLI("config validate"),
		toolname.CLI("config get session.auth_ttl"),
		toolname.CLI("config set deepseek.model deepseek-chat"),
	); err != nil {
		return err
	}
	return b.Finish("config")
}

func formatConfigSubcommands() string {
	var b strings.Builder
	for _, e := range subregistry.Entries() {
		switch e.Name {
		case "set":
			fmt.Fprintf(&b, "  %s <key> <value>   %s\n", e.Name, e.Short)
		case "get":
			fmt.Fprintf(&b, "  %s <key>           %s\n", e.Name, e.Short)
		default:
			fmt.Fprintf(&b, "  %-18s %s\n", e.Name, e.Short)
		}
	}
	return b.String()
}

// WriteConfigSetHelp delegates to the set subcommand help (tests and compat).
func WriteConfigSetHelp(w io.Writer) error {
	return set.WriteHelp(w)
}

// WriteConfigGetHelp delegates to the get subcommand help.
func WriteConfigGetHelp(w io.Writer) error {
	return cfgget.WriteHelp(w)
}

// WriteConfigInitHelp delegates to the init subcommand help.
func WriteConfigInitHelp(w io.Writer) error {
	return cfginit.WriteHelp(w)
}

// WriteConfigViewHelp delegates to the view subcommand help.
func WriteConfigViewHelp(w io.Writer) error {
	return view.WriteHelp(w)
}

// WriteConfigEditHelp delegates to the edit subcommand help.
func WriteConfigEditHelp(w io.Writer) error {
	return edit.WriteHelp(w)
}

// WriteConfigValidateHelp delegates to the validate subcommand help.
func WriteConfigValidateHelp(w io.Writer) error {
	return validate.WriteHelp(w)
}
