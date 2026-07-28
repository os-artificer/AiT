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

// Package queryflags binds the log query flags shared by the default `aitlog`
// route and `aitlog search`, and turns them into a logsrc.Query.
package queryflags

import (
	"fmt"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/logsrc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Flags holds the shared query flag values.
type Flags struct {
	Regex         string
	CaseSensitive bool
	Lines         int
	Head          bool
	Since         string
	Until         string
	Priority      string
	Unit          string
	Source        string
	Follow        bool
	JSON          bool
	Raw           bool
	UTC           bool
	NoPager       bool
	Backend       string
	Quiet         bool
	ShowVersion   bool
}

// Bind attaches the shared query flags to cmd.
func Bind(cmd *cobra.Command, f *Flags) {
	cmd.Flags().StringVarP(&f.Regex, "regex", "e", "", "match message with regular expression")
	cmd.Flags().BoolVar(&f.CaseSensitive, "case-sensitive", false, "case-sensitive keyword/regex match")
	cmd.Flags().IntVarP(&f.Lines, "lines", "n", logsrc.DefaultLimit, "last N matching records")
	cmd.Flags().BoolVar(&f.Head, "head", false, "take first N matches instead of last N")
	cmd.Flags().StringVar(&f.Since, "since", "", "show entries not older than (15m, 2h, today, RFC3339)")
	cmd.Flags().StringVar(&f.Until, "until", "", "show entries not newer than")
	cmd.Flags().StringVarP(&f.Priority, "priority", "p", "", "minimum priority (emerg…debug)")
	cmd.Flags().StringVarP(&f.Unit, "unit", "u", "", "filter by unit/service name")
	cmd.Flags().StringVarP(&f.Source, "source", "S", "", "logical source name or file path/glob")
	cmd.Flags().BoolVarP(&f.Follow, "follow", "f", false, "follow new entries (implies --no-pager)")
	cmd.Flags().BoolVar(&f.JSON, "json", false, "emit NDJSON to stdout")
	cmd.Flags().BoolVar(&f.Raw, "raw", false, "include original line in JSON output")
	cmd.Flags().BoolVar(&f.UTC, "utc", false, "format timestamps in UTC")
	cmd.Flags().BoolVar(&f.NoPager, "no-pager", false, "disable interactive pager")
	cmd.Flags().StringVar(&f.Backend, "backend", "auto", "auto|native|builtin")
	cmd.Flags().BoolVar(&f.Quiet, "quiet", false, "suppress notices on stderr")
}

// BindVersion attaches -V/--version. Only the top-level route accepts it, so
// `aitlog search version` searches for the word "version".
func BindVersion(cmd *cobra.Command, f *Flags) {
	cmd.Flags().BoolVarP(&f.ShowVersion, "version", "V", false, "print build version and exit")
}

// BuildQuery converts flags and positional keywords into a logsrc.Query, with
// now as the reference point for relative --since/--until values.
//
// Returns an error carrying Try hints for a non-positive -n, an unknown
// --backend, an unparsable priority or time expression, or a --since that is
// later than --until.
func BuildQuery(f Flags, keywords []string, now time.Time) (logsrc.Query, error) {
	q := logsrc.Query{
		Keywords:   keywords,
		Regex:      f.Regex,
		IgnoreCase: !f.CaseSensitive,
		Unit:       f.Unit,
		Source:     f.Source,
		Limit:      f.Lines,
		Head:       f.Head,
		Follow:     f.Follow,
		UTC:        f.UTC,
		Raw:        f.Raw,
		Quiet:      f.Quiet,
		MinLevel:   logsrc.LevelUnset,
	}
	if f.Lines < 1 {
		return q, errhint.WithTry(
			fmt.Errorf("%w: %d", logsrc.ErrBadLimit, f.Lines),
			"-n must be a positive number of records.",
			toolname.AILog+" -n 50",
			toolname.AILog+" --head -n 200",
		)
	}

	backend, err := parseBackend(f.Backend)
	if err != nil {
		return q, err
	}
	q.Backend = backend

	if err := applyLevel(&q, f.Priority); err != nil {
		return q, err
	}
	if err := applyRange(&q, f, now); err != nil {
		return q, err
	}
	return q, nil
}

// UsePager reports whether interactive paging should be enabled. Following and
// machine-readable output always stream.
func UsePager(f Flags) bool {
	return !f.Follow && !f.NoPager && !f.JSON
}

func parseBackend(name string) (logsrc.Backend, error) {
	switch name {
	case "", string(logsrc.BackendAuto):
		return logsrc.BackendAuto, nil
	case string(logsrc.BackendNative):
		return logsrc.BackendNative, nil
	case string(logsrc.BackendBuiltin):
		return logsrc.BackendBuiltin, nil
	default:
		return logsrc.BackendAuto, errhint.WithTry(
			fmt.Errorf("%w: %s", logsrc.ErrBadBackend, name),
			"Backend must be auto, native, or builtin.",
			toolname.AILog+" --backend auto -n 50",
			toolname.AILog+" --help",
		)
	}
}

func applyLevel(q *logsrc.Query, priority string) error {
	if priority == "" {
		return nil
	}
	level, err := logsrc.ParseLevel(priority)
	if err != nil {
		return errhint.WithTry(err, "Unrecognized log priority.",
			toolname.AILog+" -p err -n 50",
			toolname.AILog+" --help",
		)
	}
	q.MinLevel = level
	return nil
}

// applyRange parses --since/--until and rejects an inverted range, which would
// otherwise silently match nothing.
func applyRange(q *logsrc.Query, f Flags, now time.Time) error {
	if f.Since != "" {
		since, err := logsrc.ParseTimeExpression(f.Since, now)
		if err != nil {
			return errhint.WithTry(err, "Could not parse --since.",
				toolname.AILog+" --since 1h -n 50",
				toolname.AILog+" --since today",
			)
		}
		q.Since = since
	}
	if f.Until != "" {
		until, err := logsrc.ParseTimeExpression(f.Until, now)
		if err != nil {
			return errhint.WithTry(err, "Could not parse --until.",
				toolname.AILog+" --until now -n 50",
			)
		}
		q.Until = until
	}
	if q.Since.IsZero() || q.Until.IsZero() || !q.Since.After(q.Until) {
		return nil
	}
	return errhint.WithTry(
		fmt.Errorf("%w: --since is later than --until", logsrc.ErrBadSince),
		"The time range is empty.",
		toolname.AILog+" --since 2h --until 1h",
		toolname.AILog+" --since today",
	)
}
