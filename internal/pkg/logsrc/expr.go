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

import "strings"

// exprOp selects the node kind for an Expr tree.
type exprOp int

const (
	exprTerm exprOp = iota
	exprNot
	exprAnd
	exprOr
)

// Expr is a parsed --expr boolean tree over substring / phrase terms.
// Use ParseExpr to build one and Eval to test a haystack.
type Expr struct {
	op          exprOp
	text        string
	left, right *Expr
}

// Eval reports whether hay matches this expression.
// hay is the message (or raw line) to search; ignoreCase folds both hay and
// term text before substring matching. A nil Expr matches every haystack.
func (e *Expr) Eval(hay string, ignoreCase bool) bool {
	if e == nil {
		return true
	}
	if ignoreCase {
		hay = strings.ToLower(hay)
	}
	return e.eval(hay, ignoreCase)
}

// eval walks the tree; when ignoreCase is true, hay is already lowercased.
func (e *Expr) eval(hay string, ignoreCase bool) bool {
	switch e.op {
	case exprTerm:
		needle := e.text
		if ignoreCase {
			needle = strings.ToLower(needle)
		}
		return strings.Contains(hay, needle)
	case exprNot:
		return !e.left.eval(hay, ignoreCase)
	case exprAnd:
		return e.left.eval(hay, ignoreCase) && e.right.eval(hay, ignoreCase)
	case exprOr:
		return e.left.eval(hay, ignoreCase) || e.right.eval(hay, ignoreCase)
	default:
		return false
	}
}
