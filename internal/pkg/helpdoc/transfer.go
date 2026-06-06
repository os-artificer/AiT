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
	"fmt"
	"io"
)

const transferIntro = "Download or upload a single file via SFTP (OpenSSH-compatible flags)."
const transferIntroPut = "Upload a single file to a remote host via SFTP (OpenSSH-compatible flags)."
const transferNote = "Configuration is resolved via ssh -G; requires the OpenSSH client (ssh) on PATH."

const transferPaths = `  user@host:/remote/file              (get: local defaults to ./basename; put: local defaults to ./basename)
  user@host:/remote/file  local/path  (two arguments)
  destination  remote/path  local/path (three arguments)
  with --session: get remote [/local]; put local [/remote] on the saved session host
`

const transferFlagOrder = `  Flags and paths may appear in any order (same rules as ait login).
  Glued forms are supported: -p22, -i~/.ssh/id_ed25519
`

const transferAITFlags = `  --session NAME   use a saved session (remote path may omit user@host)
  --password P     SSH password (stored in OS keyring when using --session)
  --host H         legacy: remote host (do not combine with user@host in paths)
  --user U         legacy: same as -l
  --port N         legacy: same as -p
  --identity P     legacy: same as -i
`

// WriteTransferHelp prints get or put help; cmdName must be "get" or "put".
func WriteTransferHelp(w io.Writer, cmdName string) error {
	intro := transferIntro
	if cmdName == "put" {
		intro = transferIntroPut
	}
	b := NewBody(w)
	if err := b.WriteIntro(intro, transferNote); err != nil {
		return err
	}
	if err := b.WriteUsage(fmt.Sprintf("ait %s [flags] <paths...>", cmdName)); err != nil {
		return err
	}
	if err := b.WriteSection(SectionPaths, transferPaths); err != nil {
		return err
	}
	if err := b.WriteSection(SectionFlagOrder, transferFlagOrder); err != nil {
		return err
	}
	if err := b.WriteSection(SectionOpenSSHFlagsTitle, SectionOpenSSHFlags); err != nil {
		return err
	}
	if err := b.WriteSection(SectionAITFlags, transferAITFlags); err != nil {
		return err
	}
	notSupTitle := fmt.Sprintf("Not supported by ait %s (use the system ssh client instead)", cmdName)
	if err := b.WriteSection(notSupTitle, SectionOpenSSHNotSupported); err != nil {
		return err
	}
	examples := transferExamples(cmdName)
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish(cmdName)
}

func transferExamples(cmdName string) []string {
	if cmdName == "put" {
		return []string{
			"ait put user@host:/opt/app",
			"ait put -p22 -i~/.ssh/id ./bin/app user@host:/usr/local/bin/app",
			"ait put deploy@10.0.0.5 ./local.bin /opt/app/bin",
			"ait put ./patch.tar /tmp/patch.tar --session prod",
		}
	}
	return []string{
		"ait get user@deploy:/var/log/app.log",
		"ait get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log",
		"ait get deploy@10.0.0.5 /var/log/app.log ./app.log",
		"ait get ./app.log /var/log/app.log --session prod",
	}
}
