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

// Package subregistry collects ai provider subcommands registered from init hooks.
package subregistry

import (
	pkgsubregistry "github.com/os-artificer/AiT/internal/pkg/subregistry"

	"github.com/spf13/cobra"
)

// Command is the extension point for an `ai provider` subcommand.
type Command interface {
	pkgsubregistry.MetaCommand
	NewCobra() *cobra.Command
}

// Entry is name and short description for parent help.
type Entry = pkgsubregistry.Entry

var reg = pkgsubregistry.New[Command]("ai/provider/subregistry")

// Register adds a provider subcommand. Panics on duplicate Name.
func Register(c Command) {
	reg.Register(c)
}

// Attach adds all registered subcommands to parent in Order ascending.
func Attach(parent *cobra.Command) {
	reg.Attach(parent, func(c Command) *cobra.Command { return c.NewCobra() })
}

// Entries returns registered subcommands for ai provider --help.
func Entries() []Entry {
	return reg.Entries()
}
