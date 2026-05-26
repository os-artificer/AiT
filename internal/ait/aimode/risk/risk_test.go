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

// A20h: plain commands return Low.
func TestClassify_Low(t *testing.T) {
	if Classify("hostname") != Low {
		t.Fatal("expected low")
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

// ClassifyAit returns High for destructive ait subcommands.
func TestClassifyAit_HighPatterns(t *testing.T) {
	if ClassifyAit("session delete abc") != High {
		t.Fatal("expected high for session delete")
	}
	if ClassifyAit("login example.com") != High {
		t.Fatal("expected high for login")
	}
}

// ClassifyAit returns Low for read-only ait subcommands.
func TestClassifyAit_Low(t *testing.T) {
	if ClassifyAit("session --list") != Low {
		t.Fatal("expected low for session --list")
	}
}
