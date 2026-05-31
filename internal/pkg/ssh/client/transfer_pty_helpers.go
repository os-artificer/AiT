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
	"encoding/base64"
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

	// ptyPutChunkHeredocDelim is the heredoc delimiter for PTY put chunk scripts.
	ptyPutChunkHeredocDelim = "AIT_CHUNK_EOF"

	// ptyPutMktempPattern is the mktemp template for chain-mode put staging files.
	ptyPutMktempPattern = "/tmp/ait-put.XXXXXX"

	// ptyPutLeakHeadBytes is how many bytes of the remote file are sampled for leak detection.
	ptyPutLeakHeadBytes = 256
)

var (
	ptyTransferScriptHook = func(c *Client, stdin io.Writer, script string) error {
		return c.injectPutScript(stdin, script)
	}
	ptyTransferSettleHook = func(c *Client, stdin io.Writer) error {
		return c.settleShellAfterBurst(stdin)
	}
	ptyTransferMktempHook = func(c *Client, stdin io.Writer) (string, error) {
		return c.capturePutTempPath(stdin)
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

// buildPutChunkScript returns a multiline shell script that decodes a base64 chunk
// into tmpPath via heredoc (never redirects inject bytes into the final remote path).
func buildPutChunkScript(tmpPath, base64Chunk string) string {
	return fmt.Sprintf(
		"base64 -d >> %s <<'%s'\n%s\n%s",
		shellQuoteArg(tmpPath),
		ptyPutChunkHeredocDelim,
		base64Chunk,
		ptyPutChunkHeredocDelim,
	)
}

// buildPutMvScript returns a shell command that atomically moves tmpPath to remotePath.
func buildPutMvScript(tmpPath, remotePath string) string {
	return fmt.Sprintf("mv %s %s", shellQuoteArg(tmpPath), shellQuoteArg(remotePath))
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

// injectPutScript sends a multiline POSIX script through the PTY with history/echo suppression.
func (c *Client) injectPutScript(stdin io.Writer, body string) error {
	if stdin == nil {
		return io.ErrClosedPipe
	}

	body = strings.TrimRight(body, "\r\n")
	if body == "" {
		return nil
	}

	script := injectSpacePrefix + injectEchoOff + "\n" + body + "\n" + injectSpacePrefix + injectEchoOn + "\n"
	if _, err := io.WriteString(stdin, script); err != nil {
		return err
	}

	injectAt := time.Now()
	c.waitRemoteOutputQuietAfter(injectAt, injectQuietForClient(c, InjectOpts{SuppressEcho: true}), scrollback.CaptureMaxWait)
	return nil
}

// capturePutTempPath creates a remote temp file via mktemp and returns its path.
func (c *Client) capturePutTempPath(stdin io.Writer) (string, error) {
	out, err := c.injectCaptureRetry(stdin, "mktemp "+shellQuoteArg(ptyPutMktempPattern))
	if err != nil {
		return "", fmt.Errorf("mktemp: %w", err)
	}

	path := strings.TrimSpace(out)
	if path == "" {
		return "", fmt.Errorf("mktemp: empty path")
	}
	return path, nil
}

// removeRemoteFile best-effort deletes a remote path via inject.
func (c *Client) removeRemoteFile(stdin io.Writer, remotePath string) {
	if remotePath == "" {
		return
	}
	_ = c.Inject(stdin, "rm -f "+shellQuoteArg(remotePath), InjectOpts{SuppressEcho: true})
}

// remotePutContentHasInjectLeak reports whether sampled remote file bytes look like inject leakage.
func remotePutContentHasInjectLeak(sample []byte) bool {
	if len(sample) == 0 {
		return false
	}

	s := string(sample)
	if strings.Contains(s, "stty -echo") || strings.Contains(s, injectCaptureBegin) {
		return true
	}
	if strings.Contains(s, ptyPutChunkHeredocDelim) {
		return true
	}
	if strings.Contains(s, "base64 -d >>") && strings.Contains(s, "stty echo") {
		return true
	}
	return false
}

// checkRemotePutLeak samples the remote file head and rejects obvious inject command leakage.
func (t *PTYTransfer) checkRemotePutLeak(stdin io.Writer, remotePath string, expectedBytes int64) error {
	headBytes := ptyPutLeakHeadBytes
	if expectedBytes > 0 && expectedBytes+64 < int64(headBytes) {
		headBytes = int(expectedBytes) + 64
	}

	cmd := fmt.Sprintf("head -c %d %s | base64", headBytes, shellQuoteArg(remotePath))
	out, err := t.client.injectCaptureRetry(stdin, cmd)
	if err != nil {
		return err
	}

	out = strings.ReplaceAll(out, "\r", "")
	out = strings.ReplaceAll(out, "\n", "")
	if out == "" {
		return nil
	}

	sample, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		return nil
	}

	if remotePutContentHasInjectLeak(sample) {
		return fmt.Errorf("%w", ErrInjectLeakSignature)
	}

	if expectedBytes > 0 && int64(len(sample)) > expectedBytes*2 {
		return fmt.Errorf("%w: remote head larger than expected payload", ErrInjectLeakSignature)
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
