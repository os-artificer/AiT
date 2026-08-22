//go:build linux

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
	"bufio"
	"context"
	"fmt"
	"os"
	"syscall"
)

const (
	// DevKmsg is the primary kernel ring buffer device.
	DevKmsg = "/dev/kmsg"
	// ProcKmsg is the legacy interface, tried when DevKmsg is absent.
	ProcKmsg = "/proc/kmsg"
	// procUptime holds seconds since boot, used to derive the boot wall time.
	procUptime = "/proc/uptime"
)

// KmsgSource reads the kernel ring buffer, the last resort on Linux when
// neither journalctl nor the text logs are readable.
type KmsgSource struct {
	// Path overrides the device to read; empty means DevKmsg with a ProcKmsg
	// fallback.
	Path string
}

// Name returns the chain name used in notices and `ailog sources`.
func (s *KmsgSource) Name() string { return SourceKmsg }

// SupportsReverse reports false: the kernel ring buffer only reads forward.
func (s *KmsgSource) SupportsReverse() bool { return false }

// Open opens the kernel ring buffer for reading.
//
// When q.Follow is false the device is opened O_NONBLOCK so draining the
// current ring returns EOF instead of blocking forever for new messages.
// Returns ErrPermission when the device exists but is not readable (reading
// /dev/kmsg usually needs root), or the os error otherwise.
func (s *KmsgSource) Open(ctx context.Context, q Query) (Cursor, error) {
	f, path, err := openKmsg(s.Path, q.Follow)
	if err != nil {
		return nil, err
	}
	return &kmsgCursor{
		ctx:    ctx,
		f:      f,
		sc:     bufio.NewScanner(f),
		boot:   bootTime(),
		path:   path,
		follow: q.Follow,
	}, nil
}

// openKmsg opens path, falling back from DevKmsg to ProcKmsg when the former is
// missing. It reports the path actually opened. Non-follow opens are
// non-blocking so CollectLastN can finish after the ring is drained.
func openKmsg(path string, follow bool) (*os.File, string, error) {
	if path == "" {
		path = DevKmsg
	}

	flag := os.O_RDONLY
	if !follow {
		flag |= syscall.O_NONBLOCK
	}

	f, err := os.OpenFile(path, flag, 0)
	if err != nil && os.IsNotExist(err) && path == DevKmsg {
		path = ProcKmsg
		f, err = os.OpenFile(path, flag, 0)
	}
	if err == nil {
		return f, path, nil
	}

	if os.IsPermission(err) {
		return nil, path, fmt.Errorf("%w: %s: %v", ErrPermission, path, err)
	}
	return nil, path, err
}
