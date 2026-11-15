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

package net

import (
	"context"
	"io"
	"time"

	"github.com/os-artificer/AiT/internal/host/shared"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/spf13/cobra"
)

// RunConnections lists connections once, or watches until ctx is cancelled.
func RunConnections(ctx context.Context, w io.Writer, cmd *cobra.Command, f netFlags) error {
	q, err := f.connQuery(cmd)
	if err != nil {
		return err
	}
	if !f.Watch {
		return writeConnSample(ctx, w, q, f.JSON, false)
	}
	return shared.WatchEvery(ctx, f.Interval, func() error {
		// Human watch samples use --- like traffic; JSON watch has no separator.
		return writeConnSample(ctx, w, q, f.JSON, !f.JSON)
	})
}

func writeConnSample(ctx context.Context, w io.Writer, q hoststats.ConnQuery, asJSON, humanSep bool) error {
	conns, err := hoststats.ListConnections(ctx, q)
	if err != nil {
		return err
	}
	if asJSON {
		return hoststats.WriteConnectionsJSON(w, time.Now(), conns)
	}
	if _, err := io.WriteString(w, hoststats.FormatConnTable(conns)); err != nil {
		return err
	}
	if !humanSep {
		return nil
	}
	_, err = io.WriteString(w, shared.WatchSeparator)
	return err
}
