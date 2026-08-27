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

package risk

var interactiveCommands = map[string]struct{}{
	"vim": {}, "vi": {}, "nvim": {}, "nano": {}, "emacs": {},
	"less": {}, "more": {}, "top": {}, "htop": {}, "btop": {},
	"watch": {}, "man": {}, "ssh": {}, "mysql": {}, "psql": {},
	"tmux": {}, "screen": {},
}

// readCommands are observably read-only shell commands that can be
// auto-approved without user confirmation in Agent mode.
// The key is the first token of the command (lowercased).
// Commands listed in interactiveCommands take priority (e.g. less, more).
var readCommands = map[string]struct{}{
	// file listing / inspection
	"ls": {}, "cat": {}, "head": {}, "tail": {},
	"find": {}, "locate": {}, "file": {}, "stat": {}, "wc": {}, "md5sum": {},
	"sha256sum": {}, "sha1sum": {},
	// search / filter (pure read; sed/awk/tee omitted — they can write)
	"grep": {}, "egrep": {}, "fgrep": {}, "rg": {},
	"sort": {}, "uniq": {}, "cut": {}, "tr": {}, "paste": {}, "column": {},
	// system info
	"which": {}, "whereis": {}, "whoami": {}, "hostname": {}, "hostnamectl": {},
	"uname": {}, "date": {}, "uptime": {}, "id": {}, "w": {}, "who": {},
	"last": {}, "lsb_release": {}, "lscpu": {}, "lsblk": {}, "lspci": {},
	"lsusb": {}, "free": {}, "df": {}, "du": {}, "ps": {}, "pwd": {},
	"echo": {}, "env": {}, "printenv": {}, "arch": {}, "nproc": {},
	// network info (read-only queries)
	"ip": {}, "ifconfig": {}, "netstat": {}, "ss": {}, "dig": {}, "nslookup": {},
	"host": {}, "ping": {}, "traceroute": {}, "arp": {}, "route": {},
	// misc read-only
	"basename": {}, "dirname": {}, "realpath": {}, "readlink": {},
	"diff": {}, "tree": {}, "jq": {}, "yq": {}, "type": {}, "command": {},
	// git — only certain subcommands are auto-approved (see readGitSubcommands)
	"git": {},
}

// readGitSubcommands are git subcommands that are safe to auto-approve.
var readGitSubcommands = map[string]struct{}{
	"status": {}, "log": {}, "diff": {}, "show": {}, "branch": {},
	"remote": {}, "tag": {}, "stash": {}, "blame": {}, "shortlog": {},
	"describe": {}, "reflog": {}, "rev-parse": {}, "ls-files": {},
	"ls-remote": {}, "ls-tree": {}, "whatchanged": {}, "name-rev": {},
}

// highRiskFragments are destructive or privileged fragments matched per command
// segment. Fragments that span a segment operator belong in
// pipeHighRiskFragments instead, because the operator is consumed as a
// separator and can never appear inside a segment.
var highRiskFragments = []string{
	"rm -rf", "rm -fr", "mkfs", "dd if=", "shutdown", "reboot", "poweroff",
	"halt", ":(){", "chmod -r 777", "chmod 777", "> /dev/",
	"curl ", "wget ", "userdel", "passwd ", "chown -r",
}

// pipeHighRiskFragments pipe untrusted output straight into a shell
// interpreter. They straddle a segment operator, so Classify matches them
// against the whole command line.
var pipeHighRiskFragments = []string{
	"| bash", "|bash", "| sh", "|sh", "| zsh", "|zsh",
}

var highRiskAitFragments = []string{
	"session delete", "session --delete", "login ", "config edit",
}
