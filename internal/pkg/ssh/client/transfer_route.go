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

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/lrzsz"
	"github.com/os-artificer/AiT/internal/pkg/transfer/route"
	zmodem "github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

type transferInstallOpts struct {
	AssumeYes bool
	NoInstall bool
	In        io.Reader
}

// transferInstallOptsFromClient reads session-level install flags (overridable in tests).
var transferInstallOptsFromClient = func(_ *Client) transferInstallOpts {
	return transferInstallOpts{}
}

// transferInstallConfirmHook is the install confirm func (overridable in tests).
var transferInstallConfirmHook lrzsz.ConfirmFunc

// planTransfer decides ZMODEM vs SFTP and whether the target is chain-only.
func (c *Client) planTransfer(stdin io.Writer) (route.Plan, string, lrzsz.Target, error) {
	if planTransferForTest != nil {
		return planTransferForTest(c, stdin)
	}

	plan := route.Plan{}
	activeHost := ""

	if c.ChainMode() {
		plan.ChainOnly = true
	} else if c.scroll == nil {
		plan.Backend = route.BackendSFTP
		return plan, activeHost, lrzsz.TargetUnknown, nil
	} else {
		host, probeErr := c.probeActiveShellHost(stdin)
		if probeErr != nil {
			plan.Backend = route.BackendSFTP
			return plan, activeHost, lrzsz.TargetUnknown, nil
		}
		activeHost = host

		if activeShellNeedsChain(c.opts.Host, c.opts.LoginDestination, activeHost) {
			c.SetChainMode(true)
			plan.ChainOnly = true
			plan.AutoChain = true
			autoChainNotice(activeHost, c.opts.Host)
		}
	}

	if c.scroll == nil && !plan.ChainOnly {
		plan.Backend = route.BackendSFTP
		return plan, activeHost, lrzsz.TargetUnknown, nil
	}

	ok, err := c.probeLrzsz(stdin)
	if err != nil {
		return plan, activeHost, lrzsz.TargetUnknown, err
	}
	plan.LrzszOK = ok

	if ok && plan.ChainOnly {
		plan.Backend = route.BackendZModem
		return plan, activeHost, lrzsz.TargetUnknown, nil
	}

	if plan.ChainOnly {
		target, detectErr := c.detectLrzszTarget(stdin)
		if detectErr != nil {
			return plan, activeHost, lrzsz.TargetUnknown, detectErr
		}
		plan.Backend = route.BackendZModem
		return plan, activeHost, target, nil
	}

	plan.Backend = route.BackendSFTP
	return plan, activeHost, lrzsz.TargetUnknown, nil
}

func (c *Client) probeLrzsz(stdin io.Writer) (bool, error) {
	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	out, err := c.injectCaptureRetry(stdin, lrzsz.ProbeCommand())
	if err != nil {
		return false, err
	}
	return lrzsz.ParseProbeOK(out), nil
}

func (c *Client) detectLrzszTarget(stdin io.Writer) (lrzsz.Target, error) {
	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	out, err := c.injectCaptureRetry(stdin, lrzsz.DetectCommand())
	if err != nil {
		return lrzsz.TargetUnknown, err
	}
	return lrzsz.ParseTarget(out), nil
}

func (c *Client) ensureLrzszForChain(ctx context.Context, stdin io.Writer, activeHost string, target lrzsz.Target) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	opts := transferInstallOptsFromClient(c)
	installOpts := lrzsz.InstallOptions{
		AssumeYes:  opts.AssumeYes,
		NoInstall:  opts.NoInstall,
		ActiveHost: activeHost,
		In:         os.Stdin,
		Out:        os.Stderr,
		Err:        os.Stderr,
	}
	if opts.In != nil {
		installOpts.In = opts.In
	}

	allowed, err := lrzsz.ResolveInstallAllowed(true, installOpts, transferInstallConfirmHook)
	if err != nil {
		if errors.Is(err, lrzsz.ErrNonInteractiveInstall) {
			return lrzsz.WrapNonInteractive(target)
		}
		return err
	}

	if !allowed {
		return lrzsz.WrapInstallDeclined(activeHost)
	}

	cmd := lrzsz.InstallCommand(target)
	if cmd == "" {
		return lrzsz.WrapMissingChain(activeHost, target)
	}

	if err := c.runRemoteInstall(stdin, cmd); err != nil {
		return lrzsz.WrapInstallFailed(err, target)
	}

	ok, err := c.probeLrzsz(stdin)
	if err != nil {
		return lrzsz.WrapInstallFailed(err, target)
	}
	if !ok {
		return lrzsz.WrapInstallFailed(fmt.Errorf("sz/rz still missing after install"), target)
	}

	return nil
}

func (c *Client) runRemoteInstall(stdin io.Writer, cmd string) error {
	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	return c.Inject(stdin, cmd, InjectOpts{SuppressEcho: true})
}

func sftpFallbackNotice(host string) {
	short := normalizeHost(host)
	if short == "" {
		short = strings.TrimSpace(host)
	}
	if short == "" {
		short = "remote host"
	}
	aitModeNotice(
		fmt.Sprintf("lrzsz not found on %s; using SFTP.", short),
	)
}

func zmodemSftpFallbackNotice(activeHost, dialHost string) {
	_ = activeHost
	_ = dialHost
	aitModeNotice("ZMODEM failed; using SFTP.")
}

func zmodemFallbackEligible(err error) bool {
	return errors.Is(err, zmodem.ErrHandshakeTimeout) ||
		errors.Is(err, zmodem.ErrNoProtocolResponse)
}
