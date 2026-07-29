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

package logsrc_test

import (
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

func TestBuildLogShowPredicate_KeywordsAND(t *testing.T) {
	q := logsrc.Query{
		Keywords:   []string{"Failed", "password"},
		IgnoreCase: true,
	}
	got := logsrc.BuildLogShowPredicate(q)
	want := `eventMessage CONTAINS[c] "Failed" AND eventMessage CONTAINS[c] "password"`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildLogShowPredicate_EscapesQuotes(t *testing.T) {
	q := logsrc.Query{Keywords: []string{`say "hi"`}, IgnoreCase: false}
	got := logsrc.BuildLogShowPredicate(q)
	if !strings.Contains(got, `CONTAINS "say \"hi\""`) {
		t.Fatalf("%q", got)
	}
}

func TestBuildLogShowArgs_IncludesPredicateAndLast(t *testing.T) {
	q := logsrc.Query{Keywords: []string{"nosuchcmd"}, IgnoreCase: true}
	args := logsrc.BuildLogShowArgs(q, time.Hour, true)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--last 1h") {
		t.Fatalf("missing --last: %v", args)
	}
	if !strings.Contains(joined, "--predicate") {
		t.Fatalf("missing --predicate: %v", args)
	}
	if !strings.Contains(joined, `eventMessage CONTAINS[c] "nosuchcmd"`) {
		t.Fatalf("predicate body: %v", args)
	}
}

func TestBuildLogShowPredicate_ExprBoolean(t *testing.T) {
	q := logsrc.Query{Expr: `(a|b) & !c`, IgnoreCase: true}
	got := logsrc.BuildLogShowPredicate(q)
	want := `((eventMessage CONTAINS[c] "a" OR eventMessage CONTAINS[c] "b") AND NOT (eventMessage CONTAINS[c] "c"))`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildLogStreamArgs_IncludesPredicate(t *testing.T) {
	q := logsrc.Query{Keywords: []string{"x"}, IgnoreCase: true}
	args := logsrc.BuildLogStreamArgs(q)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--predicate") {
		t.Fatalf("%v", args)
	}
}

func TestBuildLogShowArgs_PriorityUsesInfoDebugNotLevel(t *testing.T) {
	errArgs := logsrc.BuildLogShowArgs(logsrc.Query{MinLevel: logsrc.LevelError}, time.Hour, true)
	joined := strings.Join(errArgs, " ")
	if strings.Contains(joined, "--level") {
		t.Fatalf("macOS log show has no --level: %v", errArgs)
	}
	if strings.Contains(joined, "--info") || strings.Contains(joined, "--debug") {
		t.Fatalf("error floor should not widen verbosity: %v", errArgs)
	}

	infoArgs := logsrc.BuildLogShowArgs(logsrc.Query{MinLevel: logsrc.LevelInfo}, time.Hour, true)
	infoJoined := strings.Join(infoArgs, " ")
	if !strings.Contains(infoJoined, "--info") || strings.Contains(infoJoined, "--debug") {
		t.Fatalf("%v", infoArgs)
	}

	debugArgs := logsrc.BuildLogStreamArgs(logsrc.Query{MinLevel: logsrc.LevelDebug})
	debugJoined := strings.Join(debugArgs, " ")
	if !strings.Contains(debugJoined, "--info") || !strings.Contains(debugJoined, "--debug") {
		t.Fatalf("%v", debugArgs)
	}
	if strings.Contains(debugJoined, "--level") {
		t.Fatalf("%v", debugArgs)
	}
}

func TestBuildJournalctlArgs_GrepKeywords(t *testing.T) {
	q := logsrc.Query{Keywords: []string{"error", "disk"}, IgnoreCase: true}
	args := logsrc.BuildJournalctlArgs(q, true, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-g ") {
		t.Fatalf("missing -g: %v", args)
	}
	if strings.Contains(joined, "--case-sensitive") {
		t.Fatalf("IgnoreCase should omit --case-sensitive: %v", args)
	}
	if strings.Contains(joined, "-n ") {
		t.Fatal("must not pass record cap")
	}
}

func TestBuildJournalctlArgs_CaseSensitive(t *testing.T) {
	q := logsrc.Query{Regex: "Error", IgnoreCase: false}
	args := logsrc.BuildJournalctlArgs(q, false, false)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-g Error") {
		t.Fatalf("%v", args)
	}
	if !strings.Contains(joined, "--case-sensitive") {
		t.Fatalf("%v", args)
	}
}

func TestBuildJournalctlArgs_SkipsGrepForExpr(t *testing.T) {
	q := logsrc.Query{Expr: "error|fail", IgnoreCase: true}
	args := logsrc.BuildJournalctlArgs(q, true, false)
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "-g ") {
		t.Fatalf("expr must not push -g: %v", args)
	}
}

func TestLogShowWindows_StartsNarrow(t *testing.T) {
	wins := logsrc.LogShowWindows
	if len(wins) < 2 || wins[0] != 5*time.Minute {
		t.Fatalf("%v", wins)
	}
}

func TestLogShowWindowsFiltered_CapsAtOneHour(t *testing.T) {
	src := &logsrc.LogShowSource{}
	wins := src.WindowsFor(logsrc.Query{Keywords: []string{"x"}, IgnoreCase: true})
	if len(wins) == 0 || wins[len(wins)-1] != time.Hour {
		t.Fatalf("%v", wins)
	}
	unfiltered := src.WindowsFor(logsrc.Query{})
	if len(unfiltered) <= len(wins) {
		t.Fatalf("unfiltered should widen further: filtered=%v unfiltered=%v", wins, unfiltered)
	}
}
