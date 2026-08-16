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

import "errors"

var (
	// ErrNoSource means every resolution layer failed or none matched.
	ErrNoSource = errors.New("no log source available")
	// ErrToolMissing means a required native tool was not found on PATH.
	ErrToolMissing = errors.New("log tool not found")
	// ErrPermission means the source exists but access was denied.
	ErrPermission = errors.New("permission denied reading logs")
	// ErrFollowUnsupported means --follow is not supported for this backend.
	ErrFollowUnsupported = errors.New("follow not supported")
	// ErrBadSince means --since/--until could not be parsed.
	ErrBadSince = errors.New("invalid time expression")
	// ErrBadLevel means -p/--priority could not be parsed.
	ErrBadLevel = errors.New("invalid log priority")
	// ErrBadBackend means --backend is not auto|native|builtin.
	ErrBadBackend = errors.New("invalid backend")
	// ErrUnknownSource means -S named a source that is not in the chain map.
	ErrUnknownSource = errors.New("unknown log source")
	// ErrBadLimit means -n was not a positive number.
	ErrBadLimit = errors.New("invalid record limit")
	// ErrEmptyRecord means a backend document held no usable record and the
	// cursor should skip it.
	ErrEmptyRecord = errors.New("empty log record")
	// ErrBadExpr means --expr could not be parsed.
	ErrBadExpr = errors.New("invalid match expression")
)
