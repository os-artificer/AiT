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

// Package registry collects top-level aihost subcommands via pkg/subregistry.
package registry

import (
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgsubregistry "github.com/os-artificer/AiT/internal/pkg/subregistry"

	"github.com/spf13/cobra"
)

// Command is a top-level aihost subcommand.
type Command interface {
	pkgsubregistry.MetaCommand
	NewCobra() *cobra.Command
}

var reg = pkgsubregistry.New[Command]("aihost/registry")

// Register adds a top-level command. Panics on duplicate Name.
func Register(c Command) {
	reg.Register(c)
}

// AttachRoot adds all registered commands to root in Order ascending.
func AttachRoot(root *cobra.Command) {
	reg.Attach(root, func(c Command) *cobra.Command { return c.NewCobra() })
}

// IsRegistered reports whether name is a registered top-level subcommand.
func IsRegistered(name string) bool {
	for _, e := range reg.Entries() {
		if e.Name == name {
			return true
		}
	}
	return false
}

// RootHelpEntries converts registered commands to root help entries.
func RootHelpEntries() []pkghelpdoc.RootCommandEntry {
	entries := reg.Entries()
	out := make([]pkghelpdoc.RootCommandEntry, len(entries))
	for i, e := range entries {
		out[i] = pkghelpdoc.RootCommandEntry{Name: e.Name, Short: e.Short}
	}
	return out
}

// Entries returns registered command metadata (for tests).
func Entries() []pkgsubregistry.Entry {
	return reg.Entries()
}
