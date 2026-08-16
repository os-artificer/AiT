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

package hoststats

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/logsrc"
)

const (
	procLogLookback       = 5 * time.Minute
	procLogCollectLimit   = 400
	procLogMinInterval    = 5 * time.Second
	procLogCollectTimeout = 2 * time.Second
	procLogServiceSuffix  = ".service"
)

// attachProcLogs fills LogLine and Logs from recs. PID matches win; otherwise
// an exact case-insensitive process-name match against Record.Process/Unit.
func attachProcLogs(procs []ProcStat, recs []logsrc.Record) {
	if len(procs) == 0 || len(recs) == 0 {
		return
	}

	byPID := map[int32][]logsrc.Record{}
	byName := map[string][]logsrc.Record{}
	for _, rec := range recs {
		if rec.PID > 0 {
			pid := int32(rec.PID)
			byPID[pid] = append(byPID[pid], rec)
		}
		if key := procLogNameKey(rec); key != "" {
			byName[key] = append(byName[key], rec)
		}
	}

	for i := range procs {
		matched := byPID[procs[i].PID]
		if len(matched) == 0 {
			matched = byName[strings.ToLower(procs[i].Name)]
		}
		logs := lastProcLogs(matched, ProcLogDetailMax)
		if len(logs) == 0 {
			continue
		}
		procs[i].Logs = logs
		procs[i].LogLine = logs[len(logs)-1].Message
	}
}

func lastProcLogs(recs []logsrc.Record, n int) []ProcLog {
	if n < 1 || len(recs) == 0 {
		return nil
	}
	if len(recs) > n {
		recs = recs[len(recs)-n:]
	}
	out := make([]ProcLog, 0, len(recs))
	for _, r := range recs {
		out = append(out, ProcLog{Time: r.Time, Message: flattenLogMessage(r.Message)})
	}
	return out
}

func procLogNameKey(rec logsrc.Record) string {
	name := rec.Process
	if name == "" {
		name = rec.Unit
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.TrimSuffix(name, procLogServiceSuffix)
	return strings.ToLower(filepath.Base(name))
}

func collectSystemLogs(ctx context.Context) ([]logsrc.Record, error) {
	r := logsrc.DefaultResolver()
	r.Notice = io.Discard
	q := logsrc.Query{
		Limit: procLogCollectLimit,
		Since: time.Now().Add(-procLogLookback),
		Quiet: true,
	}
	src, cur, err := r.OpenResolved(ctx, q)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	filt, err := logsrc.NewFilter(q)
	if err != nil {
		return nil, err
	}
	got, err := logsrc.Collect(ctx, src, cur, q, filt)
	if err != nil {
		return nil, err
	}
	return got.Records, nil
}
