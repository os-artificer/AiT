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

// Package cliroute centralizes aissh argv routing (top-level SSH default, REPL connect, meta chain).
package cliroute

import (
	"sort"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// replOnlyTopLevelWords are shell REPL builtins that must not trigger top-level SSH connect (Y10).
var replOnlyTopLevelWords = map[string]struct{}{
	"ll": {},
}

// removedTopLevelWords are legacy top-level verbs (S1/R17/LC) routed to Cobra unknown, not SSH dial.
var removedTopLevelWords = map[string]struct{}{
	"list":    {},
	"load":    {},
	"suspend": {},
	"models":  {},
	"login":   {},
	"ai":      {},
}

var reservedSubcommands map[string]struct{}

// SetReservedSubcommandNames replaces the registry-synced top-level subcommand names.
func SetReservedSubcommandNames(names []string) {
	reservedSubcommands = make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		reservedSubcommands[name] = struct{}{}
	}
}

// ReservedSubcommandNames returns sorted registered subcommand names.
func ReservedSubcommandNames() []string {
	out := make([]string, 0, len(reservedSubcommands))
	for name := range reservedSubcommands {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// IsReservedSubcommand reports whether word is a registered aissh subcommand name.
func IsReservedSubcommand(word string) bool {
	_, ok := reservedSubcommands[word]
	return ok
}

// IsRemovedTopLevelWord reports whether word is a legacy top-level verb (not SSH connect, not PATH).
func IsRemovedTopLevelWord(word string) bool {
	_, ok := removedTopLevelWords[word]
	return ok
}

// IsSSHConnectArgs reports whether top-level argv should run the SSH connect path.
// Used only from ExecuteWriters; REPL and meta use narrower predicates.
func IsSSHConnectArgs(args []string) bool {
	if len(args) == 0 || pkghelpdoc.HelpRequested(args) {
		return false
	}
	if _, replOnly := replOnlyTopLevelWords[args[0]]; replOnly {
		return false
	}
	if IsRemovedTopLevelWord(args[0]) {
		return false
	}
	return !IsReservedSubcommand(args[0])
}

// IsMetaCLIFirstWord reports whether word is handled by meta.Execute / Cobra (not chain login).
func IsMetaCLIFirstWord(word string) bool {
	if word == "list" {
		return true
	}
	return IsReservedSubcommand(word)
}

// AliasLookup resolves saved session aliases for REPL / meta chain connect predicates.
type AliasLookup interface {
	GetSession(idOrAlias string) (*store.Session, error)
}

// ShouldREPLSSHConnect reports whether a REPL line should SSH-connect after PATH miss.
func ShouldREPLSSHConnect(words []string, lookup AliasLookup) bool {
	return narrowSSHConnect(words, lookup)
}

// IsMetaChainConnectArgs reports whether !ait argv should chain-login (N1).
// lookup may be nil; when set, saved session aliases are treated like ssh config Host names.
func IsMetaChainConnectArgs(args []string, lookup AliasLookup) bool {
	return narrowSSHConnect(args, lookup)
}

func narrowSSHConnect(words []string, lookup AliasLookup) bool {
	if len(words) == 0 || IsRemovedTopLevelWord(words[0]) {
		return false
	}

	for _, w := range words {
		if strings.HasPrefix(w, "-") {
			return true
		}
	}

	first := words[0]
	if strings.Contains(first, "@") || strings.HasPrefix(first, "ssh://") {
		return true
	}

	if aliasKnown(lookup, first) || HostConfigured(first) {
		return true
	}
	return parseLoginExplicitDestination(words)
}

func aliasKnown(lookup AliasLookup, alias string) bool {
	if lookup == nil {
		return false
	}
	_, err := lookup.GetSession(alias)
	return err == nil
}

func parseLoginExplicitDestination(words []string) bool {
	spec, err := openssh.ParseLogin(words)
	if err != nil {
		return false
	}
	dest := strings.TrimSpace(spec.Destination)
	if dest == "" {
		return false
	}
	if strings.Contains(dest, "@") || strings.HasPrefix(dest, "ssh://") {
		return true
	}
	_, _, port, err := openssh.ParseDestination(dest)
	return err == nil && port > 0
}
