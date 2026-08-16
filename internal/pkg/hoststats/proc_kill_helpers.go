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
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
)

// KillProc sends SIGTERM (force=false) or SIGKILL (force=true) to pid.
// On Windows, force=false uses Terminate and force=true uses Kill.
// Returns an error when pid is invalid or the signal/terminate call fails.
func KillProc(pid int32, force bool) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("pid %d: %w", pid, err)
	}

	if force {
		if err := p.Kill(); err != nil {
			return fmt.Errorf("SIGKILL pid %d: %w", pid, err)
		}
		return nil
	}
	if err := p.Terminate(); err != nil {
		return fmt.Errorf("SIGTERM pid %d: %w", pid, err)
	}
	return nil
}
