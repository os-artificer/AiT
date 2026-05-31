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
)

const activeShellHostCmd = "hostname -f 2>/dev/null || hostname"

// ErrTransferWrongHop is returned when SFTP targets the directly connected host but the
// requested remote path is not found there (often because the active shell is deeper).
var ErrTransferWrongHop = errhint.NewSentinel("remote file not found on directly connected host")

// ErrInjectCaptureNoMarkers is returned when PTY inject capture output lacks delimiters.
var ErrInjectCaptureNoMarkers = errors.New("inject capture: markers not found")

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

// hostsMatch reports whether dialHost and activeHost refer to the same machine.
func hostsMatch(dialHost, activeHost string) bool {
	a := normalizeHost(dialHost)
	b := normalizeHost(activeHost)
	if a == "" || b == "" {
		return a == b
	}
	return a == b
}

// activeShellNeedsChain reports whether the active PTY shell host differs from the dial host.
func activeShellNeedsChain(dialHost, activeHost string) bool {
	activeHost = strings.TrimSpace(activeHost)
	if activeHost == "" {
		return false
	}
	return !hostsMatch(dialHost, activeHost)
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
		"which base64 sha256sum",
		"!ait chain off   # if file is on directly connected host",
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

	out, err := c.InjectCapture(stdin, activeShellHostCmd)
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
