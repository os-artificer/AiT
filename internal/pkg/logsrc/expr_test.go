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
	"errors"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

func TestParseExpr_OpsAndPrecedence(t *testing.T) {
	e, err := logsrc.ParseExpr(`(error|fail) & !timeout`)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eval("error disk", true) {
		t.Fatal("error should match")
	}
	if e.Eval("error timeout", true) {
		t.Fatal("timeout should exclude")
	}
	if !e.Eval("fail now", true) {
		t.Fatal("fail should match via OR")
	}
}

func TestParseExpr_WordOps(t *testing.T) {
	e, err := logsrc.ParseExpr(`(error OR fail) AND NOT timeout`)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eval("ERROR x", true) {
		t.Fatal("case fold")
	}
	if e.Eval("fail timeout", true) {
		t.Fatal("NOT timeout")
	}
}

func TestParseExpr_QuotedPhraseAndLiteralAND(t *testing.T) {
	e, err := logsrc.ParseExpr(`"disk I/O" | "AND"`)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eval("got disk I/O failure", true) {
		t.Fatal("phrase")
	}
	if !e.Eval("word AND here", false) {
		t.Fatal(`literal "AND"`)
	}
}

func TestParseExpr_StackedNot(t *testing.T) {
	e, err := logsrc.ParseExpr(`!!error`)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eval("error", true) {
		t.Fatal("double not")
	}
}

func TestParseExpr_Empty(t *testing.T) {
	_, err := logsrc.ParseExpr("  ")
	if !errors.Is(err, logsrc.ErrBadExpr) {
		t.Fatalf("%v", err)
	}
}

func TestParseExpr_BadParen(t *testing.T) {
	_, err := logsrc.ParseExpr("(a|b")
	if !errors.Is(err, logsrc.ErrBadExpr) {
		t.Fatalf("%v", err)
	}
}

func TestParseExpr_CaseSensitiveEval(t *testing.T) {
	e, err := logsrc.ParseExpr(`Error`)
	if err != nil {
		t.Fatal(err)
	}
	if e.Eval("error", false) {
		t.Fatal("should be case-sensitive")
	}
	if !e.Eval("Error", false) {
		t.Fatal("exact case")
	}
}
