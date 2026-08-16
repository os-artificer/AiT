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

// Package shared holds flag binding helpers shared by aithost subcommands.
package shared

import (
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Flags are common output / sampling options.
type Flags struct {
	JSON     bool
	Interval time.Duration
	Watch    bool
	Limit    int
	Sort     string
}

// ProcessQuery builds a ProcessQuery from flags.
// Returns an errhint-wrapped error when --sort is invalid.
func (f Flags) ProcessQuery() (hoststats.ProcessQuery, error) {
	sortKey, err := hoststats.ParseSortKey(f.Sort)
	if err != nil {
		return hoststats.ProcessQuery{}, errhint.WithTry(err,
			"Invalid --sort value.",
			toolname.AIHost+" proc --sort cpu",
			toolname.AIHost+" proc --sort mem",
		)
	}
	return hoststats.ProcessQuery{Limit: f.Limit, Sort: sortKey}, nil
}
