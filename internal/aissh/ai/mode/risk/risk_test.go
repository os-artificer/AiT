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

import "testing"

// A20h: ClassifyRisk denylist patterns return High.
func TestClassify_HighPatterns(t *testing.T) {
	if Classify("rm -rf /tmp/x") != High {
		t.Fatal("expected high")
	}
}

// A20h: plain write commands return Low.
func TestClassify_Low(t *testing.T) {
	if Classify("cp a b") != Low {
		t.Fatal("expected low for cp")
	}
	if Classify("mkdir /tmp/dir") != Low {
		t.Fatal("expected low for mkdir")
	}
}

// Read-only commands return Read.
func TestClassify_Read(t *testing.T) {
	for _, cmd := range []string{
		"ls -la", "cat /etc/hosts", "head -n5 file.txt", "tail -f /var/log/syslog",
		"find /tmp -name '*.go'", "grep -r pattern .", "wc -l file",
		"ps aux", "pwd", "echo hello", "env", "id", "hostname",
		"df -h", "du -sh /tmp", "uname -a", "date",
	} {
		if Classify(cmd) != Read {
			t.Fatalf("expected Read for %q", cmd)
		}
	}
}

// git read-only subcommands return Read.
func TestClassify_GitRead(t *testing.T) {
	for _, cmd := range []string{
		"git status", "git log --oneline", "git diff", "git branch -a",
		"git remote -v", "git show HEAD", "git tag", "git blame file.go",
	} {
		if Classify(cmd) != Read {
			t.Fatalf("expected Read for %q", cmd)
		}
	}
}

// git write subcommands return Low (not Read, not High).
func TestClassify_GitWrite(t *testing.T) {
	for _, cmd := range []string{
		"git add .", "git commit -m x", "git push", "git checkout -b feat",
		"git reset --hard", "git clean -fd",
	} {
		lvl := Classify(cmd)
		if lvl == Read {
			t.Fatalf("expected non-Read for %q, got Read", cmd)
		}
	}
}

// IsRead covers sudo-prefixed read-only commands.
func TestIsRead_SudoPrefix(t *testing.T) {
	if !IsRead("sudo cat /etc/shadow") {
		t.Fatal("expected read for sudo cat")
	}
	if IsRead("sudo vim /etc/hosts") {
		t.Fatal("vim is not read-only")
	}
}

// A20o (formerly A20h sibling): TUI launchers return Interactive.
func TestClassify_Interactive(t *testing.T) {
	if Classify("vim /etc/hosts") != Interactive {
		t.Fatal("expected interactive")
	}
}

// IsInteractive covers sudo-prefixed TUI invocations.
func TestIsInteractive_SudoPrefix(t *testing.T) {
	if !IsInteractive("sudo vim /etc/hosts") {
		t.Fatal("expected interactive for sudo vim")
	}
	if IsInteractive("sudo hostname") {
		t.Fatal("hostname is not interactive")
	}
}

// ClassifyAit returns High for destructive ai-ssh subcommands.
func TestClassifyAit_HighPatterns(t *testing.T) {
	if ClassifyAit("session delete abc") != High {
		t.Fatal("expected high for session delete")
	}
	if ClassifyAit("login example.com") != High {
		t.Fatal("expected high for login")
	}
}

// ClassifyAit returns Low for read-only ai-ssh subcommands.
func TestClassifyAit_Low(t *testing.T) {
	if ClassifyAit("session --list") != Low {
		t.Fatal("expected low for session --list")
	}
}
