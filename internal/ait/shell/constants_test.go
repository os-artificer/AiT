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

package shell

import (
	"testing"

	"github.com/os-artificer/AiT/internal/ait/login"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// TestREPLHelpArgs_FromHelpdoc (C23) keeps shell REPL help argv aligned with helpdoc.
func TestREPLHelpArgs_FromHelpdoc(t *testing.T) {
	if pkghelpdoc.HelpArgCmd != "help" || pkghelpdoc.HelpArgH != "-h" || pkghelpdoc.HelpArgHelp != "--help" {
		t.Fatalf("helpdoc tokens changed")
	}
}

// TestReplCmdLogin_MatchesLoginName (C25) wires shell login verb to login.Name.
func TestReplCmdLogin_MatchesLoginName(t *testing.T) {
	if ReplCmdLogin != login.Name {
		t.Fatalf("ReplCmdLogin=%q login.Name=%q", ReplCmdLogin, login.Name)
	}
}
