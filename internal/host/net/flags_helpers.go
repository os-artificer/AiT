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
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/spf13/cobra"
)

func bindNetFlags(cmd *cobra.Command, f *netFlags) {
	shared.BindJSONInterval(cmd, &f.Flags)
	cmd.Flags().BoolVar(&f.Conn, "conn", false, "list network connections")
	cmd.Flags().StringVar(&f.Proto, "proto", hoststats.ConnProtoTCP, "connection proto: tcp|udp|all")
	cmd.Flags().StringVar(&f.State, "state", "", "comma-separated connection states")
	cmd.Flags().BoolVar(&f.Listen, "listen", false, "show LISTEN sockets only")
	cmd.Flags().IntVar(&f.Port, "port", 0, "match local or remote port")
	cmd.Flags().Int32Var(&f.PID, "pid", 0, "filter by process id")
	cmd.Flags().IntVar(&f.ConnLimit, "limit", hoststats.DefaultConnLimit, "max connection rows")
}
