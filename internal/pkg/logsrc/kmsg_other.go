//go:build !linux

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

// kmsgLinuxOnly explains why the source is unavailable.
const kmsgLinuxOnly = "kmsg only on linux"

// KmsgSource is the kernel ring buffer reader. It exists off Linux only so
// that the chain tables compile everywhere; Open always fails.
type KmsgSource struct {
	// Path is unused off Linux.
	Path string
}

// Name returns the chain name used in notices and `ailog sources`.
func (s *KmsgSource) Name() string { return SourceKmsg }

// SupportsReverse reports false: there is nothing to read.
func (s *KmsgSource) SupportsReverse() bool { return false }

// Open always returns ErrNoSource, because no other platform exposes the Linux
// kernel ring buffer.
func (s *KmsgSource) Open(ctx context.Context, q Query) (Cursor, error) {
	return nil, fmt.Errorf("%w: %s", ErrNoSource, kmsgLinuxOnly)
}
