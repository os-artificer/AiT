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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// OpenResolved opens the first usable layer of the chain for q.Source and
// returns the Source together with its Cursor; the caller must Close the
// Cursor. A path-like or globbed q.Source always uses the built-in file
// reader. When a native layer is skipped in BackendAuto mode, a one-line
// notice explaining the fallback is written to the Resolver's notice writer.
//
// Returns ErrBadBackend for an unknown q.Backend, ErrUnknownSource for an
// unknown source name, ErrPermission when every candidate was denied, or
// ErrNoSource when no layer could be opened. All of these carry Try hints.
func (r *Resolver) OpenResolved(ctx context.Context, q Query) (Source, Cursor, error) {
	if q.Backend == "" {
		q.Backend = BackendAuto
	}
	if err := validateBackend(q.Backend); err != nil {
		return nil, nil, err
	}

	if isPathLike(q.Source) {
		src := &FileSource{Paths: []string{q.Source}, Label: q.Source}
		cur, err := openCursor(ctx, src, q)
		if err != nil {
			return nil, nil, err
		}
		return src, cur, nil
	}

	name := q.Source
	if name == "" {
		name = DefaultSourceName()
	}
	layers, ok := Chains()[name]
	if !ok {
		return nil, nil, errhint.WithTry(
			ErrUnknownSource,
			fmt.Sprintf("Unknown log source %q.", name),
			toolname.AILog+" sources",
			toolname.AILog+" -S /var/log/syslog",
			toolname.AILog+" --help",
		)
	}
	return r.openChain(ctx, q, layers)
}

// openChain walks layers in order, collecting the reason each one was skipped
// so that the final error can explain what was tried.
func (r *Resolver) openChain(ctx context.Context, q Query, layers []Layer) (Source, Cursor, error) {
	var tried []string
	var firstPerm error

	for _, layer := range layers {
		if !layerMatchesBackend(layer, q.Backend) {
			continue
		}
		if !layer.Available || layer.Open == nil {
			tried = append(tried, describeSkip(layer))
			continue
		}

		src, cur, err := openLayer(ctx, r, layer, q)
		if err != nil {
			tried = append(tried, layer.Provider+": "+err.Error())
			if errors.Is(err, ErrPermission) && firstPerm == nil {
				firstPerm = err
			}
			continue
		}
		if layer.Kind != LayerNative && q.Backend == BackendAuto {
			WriteNotice(r.notice(), q.Quiet, "%s; using %s (%s)",
				fallbackTrigger(tried), layer.Provider, layer.Kind)
		}
		return src, cur, nil
	}
	return nil, nil, r.chainError(tried, firstPerm)
}

func (r *Resolver) chainError(tried []string, firstPerm error) error {
	if firstPerm != nil {
		return errhint.WithTry(
			firstPerm,
			"Access to the log source was denied.",
			"sudo "+toolname.AILog+" -n 50",
			toolname.AILog+" sources",
		)
	}
	return errhint.WithTry(
		ErrNoSource,
		"No usable log source. Tried: "+strings.Join(tried, "; "),
		toolname.AILog+" sources",
		toolname.AILog+" -S /path/to/file.log",
		toolname.AILog+" --backend builtin -n 50",
	)
}

// Probe returns one capability row per chain layer, ordered by source name and
// then chain order, marking the layer that q.Backend would select.
func (r *Resolver) Probe(q Query) []Capability {
	if q.Backend == "" {
		q.Backend = BackendAuto
	}
	chains := Chains()
	var caps []Capability

	for _, name := range ChainNames() {
		selected := false
		for _, layer := range chains[name] {
			avail, reason := probeLayer(r, layer, name, q.Backend)
			row := Capability{
				Source:    name,
				Backend:   string(layer.Kind),
				Provider:  layer.Provider,
				Available: avail,
				Reason:    reason,
			}
			if avail && !selected {
				row.Selected = true
				selected = true
			}
			caps = append(caps, row)
		}
	}
	return caps
}
