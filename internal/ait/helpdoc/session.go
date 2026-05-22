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

import "io"

const listIntro = "List saved SSH sessions stored in the local AiT database."
const listNote = "Output columns: id, alias, host, user, active, last connected."

const loadIntro = "Reconnect to a saved session by id or alias."
const loadArguments = `  id|alias    session identifier from ait list
`

const loadAITFlags = `  --tmux          attach remote tmux session if the session was created with --tmux
  --password P    SSH password when session auth has expired (keyring unlock)
`

const suspendIntro = "Suspend the current session from the local CLI."
const suspendNote = "Inside an active SSH session, use !ait suspend instead of ait suspend."

const sessionIntro = "Manage saved SSH sessions (switch or delete)."
const sessionSubcommands = `  switch [id|alias]   connect to another saved session
  delete <id|alias|all>   remove one session or all sessions
`

const sessionSwitchIntro = "Switch to another saved session (disconnect current if active)."
const sessionSwitchArguments = `  id|alias    optional; if omitted, AiT prompts to pick a session
`

const sessionSwitchFlags = `  --tmux          use tmux for the target session
  --password P    SSH password when auth has expired
  --force         switch even if the current session is still marked active (no scrollback save)
`

const sessionDeleteIntro = "Delete a saved session or all sessions."
const sessionDeleteArguments = `  id|alias|all    session id, alias, or the literal "all"
`

const sessionDeleteFlags = `  --force    skip confirmation when deleting
`

const sessionExampleDeleteAllForce = "ait session delete all --force"

// WriteListHelp prints ait list --help to w.
func WriteListHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(listIntro, listNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait list"); err != nil {
		return err
	}
	if err := b.WriteExamples("ait list"); err != nil {
		return err
	}
	return b.Finish("list")
}

// WriteLoadHelp prints ait load --help to w.
func WriteLoadHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(loadIntro); err != nil {
		return err
	}
	if err := b.WriteUsage("ait load <id|alias> [--tmux] [--password]"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionArguments, loadArguments); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionAITFlags, loadAITFlags); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait load myserver",
		"ait load abc12345 --tmux",
		"ait load prod --password secret",
	); err != nil {
		return err
	}
	return b.Finish("load")
}

// WriteSuspendHelp prints ait suspend --help to w.
func WriteSuspendHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(suspendIntro, suspendNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait suspend"); err != nil {
		return err
	}
	if err := b.WriteExamples("!ait suspend   # inside SSH session"); err != nil {
		return err
	}
	return b.Finish("suspend")
}

// WriteSessionHelp prints ait session --help to w.
func WriteSessionHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(sessionIntro); err != nil {
		return err
	}
	if err := b.WriteUsage("ait session"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionSubcommands, sessionSubcommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait session switch prod",
		sessionExampleDeleteAllForce,
	); err != nil {
		return err
	}
	return b.Finish("session")
}

// WriteSessionSwitchHelp prints ait session switch --help to w.
func WriteSessionSwitchHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(sessionSwitchIntro); err != nil {
		return err
	}
	if err := b.WriteUsage("ait session switch [id|alias] [--tmux] [--password] [--force]"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionArguments, sessionSwitchArguments); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionAITFlags, sessionSwitchFlags); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait session switch",
		"ait session switch prod --tmux",
		"ait session switch abc12345 --force",
	); err != nil {
		return err
	}
	return b.Finish("session switch")
}

// WriteSessionDeleteHelp prints ait session delete --help to w.
func WriteSessionDeleteHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(sessionDeleteIntro); err != nil {
		return err
	}
	if err := b.WriteUsage("ait session delete <id|alias|all> [--force]"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionArguments, sessionDeleteArguments); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionAITFlags, sessionDeleteFlags); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait session delete myserver",
		sessionExampleDeleteAllForce,
	); err != nil {
		return err
	}
	return b.Finish("session delete")
}
