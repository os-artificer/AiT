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

package logsrc

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseExpr parses an --expr string into a boolean tree.
//
// Operators: & / AND, | / OR, ! / NOT (word forms are case-insensitive).
// Precedence: NOT > AND > OR. Phrases use double quotes. A nil error and
// non-nil *Expr are returned on success.
//
// Returns ErrBadExpr (possibly wrapped) for an empty expression, unbalanced
// parentheses, an unclosed quote, or any other syntax error.
func ParseExpr(s string) (*Expr, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("%w: empty expression", ErrBadExpr)
	}

	lx := &exprLexer{s: s}
	if err := lx.advance(); err != nil {
		return nil, err
	}

	e, err := lx.parseOr()
	if err != nil {
		return nil, err
	}
	if lx.tok.kind != tokEOF {
		return nil, fmt.Errorf("%w: unexpected %s", ErrBadExpr, describeExprTok(lx.tok))
	}
	return e, nil
}

// exprToPredicate converts e into an NSPredicate fragment over eventMessage.
// containsOp is "CONTAINS" or "CONTAINS[c]".
func exprToPredicate(e *Expr, containsOp string) string {
	if e == nil {
		return ""
	}
	switch e.op {
	case exprTerm:
		return "eventMessage " + containsOp + " " + quoteNSPredicate(e.text)
	case exprNot:
		return "NOT (" + exprToPredicate(e.left, containsOp) + ")"
	case exprAnd:
		return "(" + exprToPredicate(e.left, containsOp) + " AND " +
			exprToPredicate(e.right, containsOp) + ")"
	case exprOr:
		return "(" + exprToPredicate(e.left, containsOp) + " OR " +
			exprToPredicate(e.right, containsOp) + ")"
	default:
		return ""
	}
}

func describeExprTok(t exprTok) string {
	switch t.kind {
	case tokEOF:
		return "end of expression"
	case tokTerm:
		return fmt.Sprintf("term %q", t.text)
	case tokAnd:
		return exprWordAND
	case tokOr:
		return exprWordOR
	case tokNot:
		return exprWordNOT
	case tokLParen:
		return "("
	case tokRParen:
		return ")"
	default:
		return "token"
	}
}

func isExprMeta(r rune) bool {
	return r == '&' || r == '|' || r == '!' || r == '(' || r == ')' || r == '"' || unicode.IsSpace(r)
}
