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
	"github.com/os-artificer/AiT/internal/host/shared"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// netFlags holds traffic (embedded) and connection-view options.
// ConnLimit is separate from Flags.Limit so traffic Sample stays unaffected.
type netFlags struct {
	shared.Flags
	Conn      bool
	Proto     string
	State     string
	Listen    bool
	Port      int
	PID       int32
	ConnLimit int
}

func (f netFlags) connFilterChanged(cmd *cobra.Command) bool {
	for _, name := range []string{"proto", "state", "listen", "port", "pid", "limit"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func (f netFlags) validate(cmd *cobra.Command) error {
	if f.Conn {
		return f.validateConn(cmd)
	}
	if f.connFilterChanged(cmd) {
		return errhint.WithTryMsg(
			"connection filters require --conn",
			"Use --conn to list connections, or omit connection filters for traffic.",
			toolname.HostCLI("net --conn"),
			toolname.HostCLI("net --help"),
		)
	}
	return nil
}

func (f netFlags) validateConn(cmd *cobra.Command) error {
	if f.Listen && cmd.Flags().Changed("state") && f.State != "" {
		return errhint.WithTryMsg(
			"--listen and --state are mutually exclusive",
			"Pass either --listen or --state, not both.",
			toolname.HostCLI("net --conn --listen"),
			toolname.HostCLI("net --conn --state ESTABLISHED"),
		)
	}
	if cmd.Flags().Changed("port") && (f.Port < 1 || f.Port > 65535) {
		return errhint.WithTryMsg(
			"invalid --port",
			"Port must be between 1 and 65535.",
			toolname.HostCLI("net --conn --port 22"),
		)
	}
	if cmd.Flags().Changed("pid") && f.PID <= 0 {
		return errhint.WithTryMsg(
			"invalid --pid",
			"PID must be a positive process id.",
			toolname.HostCLI("net --conn --pid 1"),
		)
	}
	if cmd.Flags().Changed("limit") && f.ConnLimit <= 0 {
		return errhint.WithTryMsg(
			"invalid --limit",
			"Connection --limit must be a positive integer.",
			toolname.HostCLI("net --conn --limit 100"),
		)
	}
	if _, err := f.connQuery(cmd); err != nil {
		return err
	}
	return nil
}

func (f netFlags) connQuery(cmd *cobra.Command) (hoststats.ConnQuery, error) {
	q := hoststats.ConnQuery{
		Proto: f.Proto,
		Port:  f.Port,
		PID:   f.PID,
		Limit: f.ConnLimit,
	}
	if f.Listen {
		q.States = []string{hoststats.ConnStatusListen}
	} else if cmd.Flags().Changed("state") && f.State != "" {
		q.States = []string{f.State}
	}
	if !cmd.Flags().Changed("proto") {
		q.Proto = hoststats.ConnProtoTCP
	}
	normalized, err := hoststats.NormalizeConnQuery(q)
	if err != nil {
		return hoststats.ConnQuery{}, errhint.WithTry(err,
			"Invalid connection filter.",
			toolname.HostCLI("net --conn --proto tcp"),
			toolname.HostCLI("net --conn --state ESTABLISHED"),
			toolname.HostCLI("net --help"),
		)
	}
	return normalized, nil
}
