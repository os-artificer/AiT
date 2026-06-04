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
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/lrzsz"
	pkgzmodem "github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

const activeShellHostCmd = "hostname -f 2>/dev/null || hostname"

// ErrTransferWrongHop is returned when SFTP targets the directly connected host but the
// requested remote path is not found there (often because the active shell is deeper).
var ErrTransferWrongHop = errhint.NewSentinel("remote file not found on directly connected host")

// ErrInjectCaptureNoMarkers is returned when PTY inject capture output lacks delimiters.
var ErrInjectCaptureNoMarkers = errors.New("inject capture: markers not found")

// ErrMissingLrzsz is returned when the remote active shell lacks lrzsz (sz/rz).
var ErrMissingLrzsz = lrzsz.ErrMissing

// normalizeHost lowercases host, strips user@ and port, and returns the first DNS label
// so short names and FQDNs compare equal (e.g. bastion and bastion.example.com).
func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}

	if i := strings.LastIndex(host, "@"); i >= 0 {
		host = host[i+1:]
	}

	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if i := strings.Index(host, "."); i >= 0 {
		host = host[:i]
	}

	return host
}

// normalizeHostFull lowercases host, strips user@ and port, keeping the full hostname.
func normalizeHostFull(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}

	if i := strings.LastIndex(host, "@"); i >= 0 {
		host = host[i+1:]
	}

	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	return host
}

// hostsMatchFullName reports whether a and b are the same host by full hostname (not first DNS label only).
func hostsMatchFullName(a, b string) bool {
	na := normalizeHostFull(a)
	nb := normalizeHostFull(b)
	if na == "" || nb == "" {
		return na == nb
	}
	return na == nb
}

// hostsMatch reports whether dialHost and activeHost refer to the same machine.
func hostsMatch(dialHost, activeHost string) bool {
	a := normalizeHost(dialHost)
	b := normalizeHost(activeHost)
	if a == "" || b == "" {
		return a == b
	}
	if a == b {
		return true
	}
	return hostsMatchFullName(dialHost, activeHost)
}

// hostsMatchIdentities reports whether activeHost refers to the dial target, comparing
// opts.Host, LoginDestination (ssh alias), and full hostname forms.
func hostsMatchIdentities(dialHost, loginDest, activeHost string) bool {
	activeHost = strings.TrimSpace(activeHost)
	if activeHost == "" {
		return false
	}

	if hostsMatch(dialHost, activeHost) {
		return true
	}

	loginDest = strings.TrimSpace(loginDest)
	if loginDest != "" && hostsMatch(loginDest, activeHost) {
		return true
	}

	return false
}

// activeShellNeedsChain reports whether the active PTY shell is on a host deeper than c.conn.
func activeShellNeedsChain(dialHost, loginDest, activeHost string) bool {
	activeHost = strings.TrimSpace(activeHost)
	if activeHost == "" {
		return false
	}
	return !hostsMatchIdentities(dialHost, loginDest, activeHost)
}

func wrapTransferWrongHop(dialHost string, cause error) error {
	host := normalizeHost(dialHost)
	if host == "" {
		host = dialHost
	}
	explanation := fmt.Sprintf(
		"remote file not found on directly connected host (%s).",
		host,
	)
	return errhint.WithTry(
		fmt.Errorf("%w: %v", ErrTransferWrongHop, cause),
		explanation,
		"!ait chain on",
		"!ait get /path ./local",
	)
}

func wrapPTYCaptureErr(cause error) error {
	return errhint.WithTry(
		cause,
		"PTY file transfer capture failed.",
		"!ait put <local-on-this-machine> <remote-on-active-shell>",
		"which base64 sha256sum",
		"!ait chain off   # if file is on directly connected host",
	)
}

func wrapPTYZModemErr(cause error) error {
	try := lrzsz.ManualTryLines(lrzsz.TargetUnknown)
	if errors.Is(cause, pkgzmodem.ErrTeardownTimeout) {
		try = append([]string{
			"If the local file size already matches the remote, retry the transfer once.",
			"On the active shell: sz --version  (upgrade lrzsz if very old)",
		}, try...)
	}
	return errhint.WithTry(
		cause,
		"Chain-mode file transfer uses ZMODEM (lrzsz) over the PTY.",
		try...,
	)
}

func isSFTPNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "file or parent directory may not exist") ||
		strings.Contains(msg, "no such file")
}

func (c *Client) probeActiveShellHost(stdin io.Writer) (string, error) {
	c.pauseRemoteOut.Store(true)
	defer c.pauseRemoteOut.Store(false)

	out, err := c.injectCaptureRetry(stdin, activeShellHostCmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func autoChainNotice(activeHost, dialHost string) {
	activeShort := normalizeHost(activeHost)
	dialShort := normalizeHost(dialHost)
	if activeShort == "" {
		activeShort = strings.TrimSpace(activeHost)
	}
	if dialShort == "" {
		dialShort = dialHost
	}
	aitModeNotice(
		fmt.Sprintf("chain mode auto-enabled: active shell is on %s (connected to %s).", activeShort, dialShort),
		"File transfer will use PTY-pipe injection.",
		"Use !ait chain off to disable.",
	)
}
