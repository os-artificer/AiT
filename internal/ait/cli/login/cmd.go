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

// Package login builds the `ait login` Cobra command (OpenSSH-compatible).
package login

import (
	"context"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	pkglogin "github.com/os-artificer/AiT/internal/ait/login"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr func(fn func(ctx context.Context, mgr *session.Manager) error) error

// New builds the `ait login` cobra command.
func New(runWithMgr RunWithMgr) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "login [flags] destination [command [argument ...]]",
		Short:                 "Log in to a host (OpenSSH-compatible)",
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceErrors:         true,
	}
	helpdoc.BindCustomArgv(cmd, helpdoc.WriteLoginHelp, func(_ *cobra.Command, argv []string) error {
		return runWithMgr(func(ctx context.Context, mgr *session.Manager) error {
			return pkglogin.Run(ctx, mgr, argv)
		})
	})
	return cmd
}
