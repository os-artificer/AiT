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

package aimode

var interactiveCommands = map[string]struct{}{
	"vim": {}, "vi": {}, "nvim": {}, "nano": {}, "emacs": {},
	"less": {}, "more": {}, "top": {}, "htop": {}, "btop": {},
	"watch": {}, "man": {}, "ssh": {}, "mysql": {}, "psql": {},
	"tmux": {}, "screen": {},
}

var highRiskFragments = []string{
	"rm -rf", "rm -fr", "mkfs", "dd if=", "shutdown", "reboot", "poweroff",
	"halt", ":(){", "chmod -r 777", "chmod 777", "> /dev/", "| bash", "| sh",
	"curl ", "| bash", "wget ", "| sh", "userdel", "passwd ", "chown -r",
}

var highRiskAitFragments = []string{
	"session delete", "session --delete", "login ", "config edit",
}
