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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/scrollback"
)

const (
	// ptyTransferChunkSize is the maximum base64 chunk size in bytes sent per
	// PTY injection. 3KB of raw data produces ~4KB of base64, which stays
	// safely within PTY line buffer limits.
	ptyTransferChunkSize = 3072

	// transferPostBurstQuiet is the idle period after a burst of PTY injects before capture.
	transferPostBurstQuiet = 400 * time.Millisecond

	// transferShellSettleCmd is a POSIX no-op run after chunk uploads before verify capture.
	transferShellSettleCmd = "true"
)

var (
	ptyTransferInjectHook = func(c *Client, stdin io.Writer, cmd string, opts InjectOpts) error {
		return c.Inject(stdin, cmd, opts)
	}
	ptyTransferVerifyHook = func(c *Client, stdin io.Writer, remotePath string) (string, error) {
		return c.settleAndCaptureHash(stdin, remotePath)
	}
)

// chunkString splits s into chunks of at most chunkSize bytes.
func chunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}

	return chunks
}

// buildPutChunkCmd constructs a shell command that decodes a base64 chunk and
// appends it to remotePath. The command uses printf to avoid heredoc edge cases
// and pipe to base64 -d for decoding.
func buildPutChunkCmd(remotePath string, base64Chunk string) string {
	return fmt.Sprintf("printf '%%s' %s | base64 -d >> %s", shellQuoteArg(base64Chunk), shellQuoteArg(remotePath))
}

// shellQuoteArg wraps s in single quotes with embedded single-quote escaping.
// This produces a POSIX-safe quoted string for shell command construction.
// Reuses the same quoting convention as BuildRemoteCommandLine in dialchain.go.
func shellQuoteArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// ptyAbsLocalPath returns the absolute path for localPath, resolving relative
// paths against the current working directory.
func ptyAbsLocalPath(localPath string) (string, error) {
	if filepath.IsAbs(localPath) {
		return localPath, nil
	}
	return filepath.Abs(localPath)
}

// localFileHash computes the sha256sum of data and returns the hex-encoded hash.
func localFileHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// verifyLocalHash reads the file at path and verifies its sha256sum matches expected.
// Returns a descriptive error on mismatch or file read failure.
func verifyLocalHash(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	actual := localFileHash(data)
	if actual != expected {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

// settleShellAfterBurst runs a POSIX no-op and waits for the remote shell to become idle.
func (c *Client) settleShellAfterBurst(stdin io.Writer) error {
	if err := c.Inject(stdin, transferShellSettleCmd, InjectOpts{SuppressEcho: true}); err != nil {
		return err
	}

	injectAt := time.Now()
	quiet := transferPostBurstQuiet
	if c.ChainMode() {
		quiet = transferChainCaptureQuiet
	}
	c.waitRemoteOutputQuietAfter(injectAt, quiet, scrollback.CaptureMaxWait)
	if c.scroll != nil {
		c.scroll.StripSubstrings(injectStripFragments()...)
	}
	return nil
}

// settleAndCaptureHash waits for the shell to settle after a burst inject, then captures sha256sum output.
func (c *Client) settleAndCaptureHash(stdin io.Writer, remotePath string) (string, error) {
	if err := c.settleShellAfterBurst(stdin); err != nil {
		return "", err
	}

	cmd := "sha256sum " + shellQuoteArg(remotePath)
	return c.injectCaptureRetry(stdin, cmd)
}
