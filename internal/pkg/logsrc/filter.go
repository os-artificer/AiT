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
	"regexp"
	"strings"
)

// Filter applies Go-side keyword/regex/expr/level/time predicates to a record.
type Filter struct {
	keywords []string
	re       *regexp.Regexp
	expr     *Expr
	q        Query
}

// NewFilter builds a Filter from q.
//
// When q.Expr is set, it is parsed once and used for text matching instead of
// Keywords and Regex. Returns a wrapped ErrBadExpr if Expr is invalid, or a
// compile error if Regex is invalid.
func NewFilter(q Query) (*Filter, error) {
	f := &Filter{q: q}
	if q.Expr != "" {
		e, err := ParseExpr(q.Expr)
		if err != nil {
			return nil, err
		}
		f.expr = e
		return f, nil
	}

	if q.IgnoreCase {
		for _, k := range q.Keywords {
			f.keywords = append(f.keywords, strings.ToLower(k))
		}
	} else {
		f.keywords = append(f.keywords, q.Keywords...)
	}

	if q.Regex != "" {
		pat := q.Regex
		if q.IgnoreCase && !strings.HasPrefix(pat, "(?") {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		f.re = re
	}
	return f, nil
}

// Match reports whether rec satisfies all filter predicates. Keywords match
// with AND semantics unless --expr is set, in which case the boolean
// expression is used instead.
//
// Unknown record attributes never exclude a record: a zero Time passes the
// --since/--until range and a LevelUnset record passes the --priority floor,
// so lines that a text backend could not fully parse still reach the user.
func (f *Filter) Match(rec Record) bool {
	if f == nil {
		return true
	}
	if !f.matchTime(rec) || !f.matchLevel(rec) || !f.matchUnit(rec) {
		return false
	}
	return f.matchText(rec)
}

// matchTime applies the --since/--until range. A record with no timestamp is
// treated as unknown and kept.
func (f *Filter) matchTime(rec Record) bool {
	if rec.Time.IsZero() {
		return true
	}
	if !f.q.Since.IsZero() && rec.Time.Before(f.q.Since) {
		return false
	}
	return f.q.Until.IsZero() || !rec.Time.After(f.q.Until)
}

// matchLevel applies the --priority floor. A LevelUnset record is treated as
// unknown severity and kept.
func (f *Filter) matchLevel(rec Record) bool {
	if f.q.MinLevel == LevelUnset || rec.Level == LevelUnset {
		return true
	}
	return rec.Level <= f.q.MinLevel
}

// matchUnit applies --unit, comparing case-insensitively.
func (f *Filter) matchUnit(rec Record) bool {
	return f.q.Unit == "" || strings.EqualFold(rec.Unit, f.q.Unit)
}

// matchText applies --expr, or --regex and the keyword AND set, to the
// message (falling back to the raw line when a backend produced no message).
func (f *Filter) matchText(rec Record) bool {
	hay := rec.Message
	if hay == "" {
		hay = rec.Raw
	}
	if f.expr != nil {
		return f.expr.Eval(hay, f.q.IgnoreCase)
	}
	if f.re != nil && !f.re.MatchString(hay) {
		return false
	}

	if f.q.IgnoreCase {
		hay = strings.ToLower(hay)
	}
	for _, k := range f.keywords {
		if !strings.Contains(hay, k) {
			return false
		}
	}
	return true
}
