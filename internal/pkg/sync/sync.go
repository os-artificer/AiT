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

package syncpkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/sshclient"
	"github.com/pkg/sftp"
)

// Options for file sync.
type Options struct {
	Src string
	Dst string
	SSH sshclient.Options
	AI  *ai.Client
}

// Run syncs files using rsync, scp, or SFTP fallback.
// Password-authenticated sessions use the in-process SFTP client because
// non-interactive rsync/scp cannot read AiT's keyring-backed password.
func Run(ctx context.Context, opts Options) error {
	if opts.SSH.Password != "" {
		return runSFTP(ctx, opts)
	}
	if _, err := platform.LookPath("rsync"); err == nil {
		return runRsync(ctx, opts)
	}
	if _, err := platform.LookPath("scp"); err == nil {
		return runScp(ctx, opts)
	}
	if err := runSFTP(ctx, opts); err == nil {
		return nil
	}
	if opts.AI != nil {
		suggestion, serr := opts.AI.SuggestInstall(ctx, "rsync")
		if serr == nil {
			return fmt.Errorf("rsync/scp not found. Suggestion:\n%s", suggestion)
		}
	}
	return fmt.Errorf("rsync and scp not found; install rsync or scp")
}

// sshArgs returns argv for an `ssh` invocation (no shell escaping).
func sshArgs(o sshclient.Options) []string {
	args := []string{"-o", "BatchMode=no"}
	if os.Getenv("AIT_SSH_INSECURE") == "1" {
		args = append(args, "-o", "StrictHostKeyChecking=no")
	} else if kh := sshclient.KnownHostsPath(); kh != "" {
		args = append(args, "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+kh)
	}
	if o.Password != "" {
		args = append(args, "-o", "PreferredAuthentications=password", "-o", "PubkeyAuthentication=no")
	}
	if o.IdentityPath != "" {
		args = append(args, "-i", o.IdentityPath)
	}
	port := o.Port
	if port == 0 {
		port = 22
	}
	args = append(args, "-p", fmt.Sprintf("%d", port))
	return args
}

// shellQuote returns a POSIX shell single-quoted form of s.
// rsync's -e argument is parsed by /bin/sh on the local machine, so any path
// passed via -i must be shell-safe even if it contains spaces or quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// rshArg builds the value for `rsync -e` with each argv element shell-quoted.
func rshArg(o sshclient.Options) string {
	parts := []string{"ssh"}
	for _, a := range sshArgs(o) {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

func runRsync(ctx context.Context, opts Options) error {
	args := []string{"-avz", "--progress", "-e", rshArg(opts.SSH), opts.Src, opts.Dst}
	cmd := exec.CommandContext(ctx, "rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runScp(ctx context.Context, opts Options) error {
	args := sshArgs(opts.SSH)
	args = append(args, "-r", opts.Src, opts.Dst)
	cmd := exec.CommandContext(ctx, "scp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runSFTP(ctx context.Context, opts Options) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	client, err := sshclient.Dial(opts.SSH)
	if err != nil {
		return err
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client.Conn())
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	user, host, remotePath, err := parseRemoteDest(opts.Dst)
	if err != nil {
		return err
	}
	_ = user
	_ = host
	f, err := os.Open(opts.Src)
	if err != nil {
		return err
	}
	defer f.Close()
	remote, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer remote.Close()
	_, err = remote.ReadFrom(f)
	return err
}

// parseRemoteDest parses user@host:/path and user@[ipv6]:/path forms.
func parseRemoteDest(dst string) (user, host, path string, err error) {
	var userHost string
	switch {
	case strings.Contains(dst, "]:"):
		i := strings.Index(dst, "]:")
		userHost = dst[:i+1]
		path = dst[i+2:]
	case strings.Contains(dst, ":/"):
		i := strings.Index(dst, ":/")
		userHost = dst[:i]
		path = dst[i+1:]
	default:
		i := strings.LastIndex(dst, ":")
		if i <= 0 {
			return "", "", "", fmt.Errorf("invalid remote path %q", dst)
		}
		userHost = dst[:i]
		path = dst[i+1:]
	}
	if at := strings.LastIndex(userHost, "@"); at >= 0 {
		user = userHost[:at]
		host = userHost[at+1:]
	} else {
		host = userHost
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	if host == "" {
		return "", "", "", fmt.Errorf("invalid remote path %q", dst)
	}
	return user, host, path, nil
}
