//go:build !windows

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
	"fmt"
)

// EvtAPISource is the Windows Event Log reader. Off Windows it exists only so
// that the package compiles on every platform, and always fails to open.
type EvtAPISource struct {
	Channel string
	Label   string
}

// Name returns the channel label.
func (s *EvtAPISource) Name() string {
	if s.Label != "" {
		return s.Label
	}
	return s.Channel
}

// SupportsReverse reports true to match the Windows implementation.
func (s *EvtAPISource) SupportsReverse() bool { return true }

// Open always returns ErrNoSource off Windows.
func (s *EvtAPISource) Open(ctx context.Context, q Query) (Cursor, error) {
	return nil, fmt.Errorf("%w: the Event Log API is only available on Windows", ErrNoSource)
}
