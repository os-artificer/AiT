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
	"unicode/utf8"
)

type exprTokKind int

const (
	tokEOF exprTokKind = iota
	tokTerm
	tokAnd
	tokOr
	tokNot
	tokLParen
	tokRParen
)

const (
	exprWordAND = "AND"
	exprWordOR  = "OR"
	exprWordNOT = "NOT"
)

type exprTok struct {
	kind exprTokKind
	text string
}

// exprLexer tokenizes and parses an --expr string.
type exprLexer struct {
	s   string
	i   int
	tok exprTok
	err error
}

func (lx *exprLexer) parseOr() (*Expr, error) {
	left, err := lx.parseAnd()
	if err != nil {
		return nil, err
	}
	for lx.tok.kind == tokOr {
		if err := lx.advance(); err != nil {
			return nil, err
		}
		right, err := lx.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &Expr{op: exprOr, left: left, right: right}
	}
	return left, nil
}

func (lx *exprLexer) parseAnd() (*Expr, error) {
	left, err := lx.parseUnary()
	if err != nil {
		return nil, err
	}
	for lx.tok.kind == tokAnd {
		if err := lx.advance(); err != nil {
			return nil, err
		}
		right, err := lx.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &Expr{op: exprAnd, left: left, right: right}
	}
	return left, nil
}

func (lx *exprLexer) parseUnary() (*Expr, error) {
	if lx.tok.kind != tokNot {
		return lx.parsePrimary()
	}
	if err := lx.advance(); err != nil {
		return nil, err
	}
	x, err := lx.parseUnary()
	if err != nil {
		return nil, err
	}
	return &Expr{op: exprNot, left: x}, nil
}

func (lx *exprLexer) parsePrimary() (*Expr, error) {
	switch lx.tok.kind {
	case tokTerm:
		e := &Expr{op: exprTerm, text: lx.tok.text}
		if err := lx.advance(); err != nil {
			return nil, err
		}
		return e, nil
	case tokLParen:
		if err := lx.advance(); err != nil {
			return nil, err
		}
		e, err := lx.parseOr()
		if err != nil {
			return nil, err
		}
		if lx.tok.kind != tokRParen {
			return nil, fmt.Errorf("%w: missing )", ErrBadExpr)
		}
		if err := lx.advance(); err != nil {
			return nil, err
		}
		return e, nil
	case tokEOF:
		return nil, fmt.Errorf("%w: unexpected end of expression", ErrBadExpr)
	default:
		return nil, fmt.Errorf("%w: unexpected %s", ErrBadExpr, describeExprTok(lx.tok))
	}
}

// advance calls next and returns any lexical error.
func (lx *exprLexer) advance() error {
	lx.next()
	return lx.err
}

// next advances to the next token, or sets lx.err on a lexical error.
func (lx *exprLexer) next() {
	if lx.err != nil {
		lx.tok = exprTok{kind: tokEOF}
		return
	}

	lx.skipSpace()
	if lx.i >= len(lx.s) {
		lx.tok = exprTok{kind: tokEOF}
		return
	}

	switch c := lx.s[lx.i]; c {
	case '(':
		lx.i++
		lx.tok = exprTok{kind: tokLParen}
		return
	case ')':
		lx.i++
		lx.tok = exprTok{kind: tokRParen}
		return
	case '&':
		lx.i++
		lx.tok = exprTok{kind: tokAnd}
		return
	case '|':
		lx.i++
		lx.tok = exprTok{kind: tokOr}
		return
	case '!':
		lx.i++
		lx.tok = exprTok{kind: tokNot}
		return
	case '"':
		text, err := lx.readQuoted()
		if err != nil {
			lx.err = err
			lx.tok = exprTok{kind: tokEOF}
			return
		}
		lx.tok = exprTok{kind: tokTerm, text: text}
		return
	}

	lx.readWordOrOp()
}

// readWordOrOp consumes an unquoted word and classifies AND/OR/NOT operators.
func (lx *exprLexer) readWordOrOp() {
	start := lx.i
	for lx.i < len(lx.s) {
		r, size := utf8.DecodeRuneInString(lx.s[lx.i:])
		if isExprMeta(r) {
			break
		}
		lx.i += size
	}
	word := lx.s[start:lx.i]
	if word == "" {
		lx.err = fmt.Errorf("%w: unexpected character", ErrBadExpr)
		lx.tok = exprTok{kind: tokEOF}
		return
	}
	switch strings.ToUpper(word) {
	case exprWordAND:
		lx.tok = exprTok{kind: tokAnd}
	case exprWordOR:
		lx.tok = exprTok{kind: tokOr}
	case exprWordNOT:
		lx.tok = exprTok{kind: tokNot}
	default:
		lx.tok = exprTok{kind: tokTerm, text: word}
	}
}

func (lx *exprLexer) skipSpace() {
	for lx.i < len(lx.s) {
		r, size := utf8.DecodeRuneInString(lx.s[lx.i:])
		if !unicode.IsSpace(r) {
			return
		}
		lx.i += size
	}
}

// readQuoted reads a double-quoted phrase starting at the opening quote.
func (lx *exprLexer) readQuoted() (string, error) {
	lx.i++ // skip opening "
	var b strings.Builder
	for lx.i < len(lx.s) {
		c := lx.s[lx.i]
		if c == '"' {
			lx.i++
			return b.String(), nil
		}
		if c == '\\' && lx.i+1 < len(lx.s) {
			b.WriteByte(lx.s[lx.i+1])
			lx.i += 2
			continue
		}
		b.WriteByte(c)
		lx.i++
	}
	return "", fmt.Errorf("%w: unclosed quote", ErrBadExpr)
}
