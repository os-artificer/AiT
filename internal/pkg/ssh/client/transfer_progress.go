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
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/errhint"

	xterm "golang.org/x/term"
)

const (
	transferProgressPrefix       = errhint.PrefixAiT
	transferProgressPctFmt       = "copying… %d%%"
	transferProgressClosingFmt   = "closing…"
	transferProgressFinishingFmt = "finishing…"
)

// transferProgressPhaseHook, when set, replaces Phase output (tests only).
var transferProgressPhaseHook func(p *TransferProgress, msg string)

// transferProgressInlineTTY, when non-nil, overrides inline TTY detection (tests only).
var transferProgressInlineTTY func(p *TransferProgress) bool

// TransferProgress writes phase and completion lines to stderr using CRLF-safe termPrintln.
type TransferProgress struct {
	w            io.Writer
	bytes        int64
	lastPct      int64
	inlineActive bool
}

// NewTransferProgress returns a progress reporter writing to w, or os.Stderr when w is nil.
func NewTransferProgress(w io.Writer) *TransferProgress {
	if w == nil {
		w = os.Stderr
	}
	return &TransferProgress{w: w}
}

// Phase prints one progress line prefixed with AiT:.
func (p *TransferProgress) Phase(msg string) {
	if p == nil {
		return
	}
	if transferProgressPhaseHook != nil {
		transferProgressPhaseHook(p, msg)
		return
	}
	termPrintln(p.w, transferProgressPrefix+msg)
}

// PhaseInline updates the current progress line in place on a TTY; otherwise prints a new line.
func (p *TransferProgress) PhaseInline(msg string) {
	if p == nil {
		return
	}
	if transferProgressPhaseHook != nil {
		transferProgressPhaseHook(p, msg)
		return
	}
	if p.inlineTTYEnabled() {
		termPrintInline(p.w, transferProgressPrefix+msg)
		p.inlineActive = true
		return
	}
	termPrintln(p.w, transferProgressPrefix+msg)
}

// Done prints the transfer completion line.
func (p *TransferProgress) Done(msg string) {
	if p == nil {
		return
	}
	if transferProgressPhaseHook != nil {
		transferProgressPhaseHook(p, msg)
		p.inlineActive = false
		return
	}
	if p.inlineActive {
		termPrintln(p.w, "\r"+transferProgressPrefix+msg)
		p.inlineActive = false
		return
	}
	p.Phase(msg)
}

// ChainModeOn reports that PTY chain mode is already active for this transfer.
func (p *TransferProgress) ChainModeOn() {
	p.Phase("transfer via PTY (chain mode on)")
}

// StartGet announces a download and records expected size when size >= 0.
func (p *TransferProgress) StartGet(remotePath, localPath string, size int64) {
	p.Phase(formatTransferStart("get", remotePath, localPath, size))
}

// StartPut announces an upload and records expected size when size >= 0.
func (p *TransferProgress) StartPut(localPath, remotePath string, size int64) {
	p.Phase(formatTransferStart("put", localPath, remotePath, size))
	if size >= 0 {
		p.bytes = size
	}
}

// reportPercent is the shared inline percent reporter for get/put, SFTP, and ZMODEM.
func (p *TransferProgress) reportPercent(pct int64) {
	if p == nil {
		return
	}
	if pct == p.lastPct {
		return
	}
	p.lastPct = pct
	p.PhaseInline(fmt.Sprintf(transferProgressPctFmt, pct))
}

// PhaseClosing reports the SFTP close phase after copy reaches 100%.
func (p *TransferProgress) PhaseClosing() {
	if p == nil {
		return
	}
	p.PhaseInline(transferProgressClosingFmt)
}

// PhaseFinishing reports the ZMODEM protocol teardown phase after copy reaches 100%.
func (p *TransferProgress) PhaseFinishing() {
	if p == nil {
		return
	}
	if transferProgressPhaseHook != nil {
		transferProgressPhaseHook(p, transferProgressFinishingFmt)
		return
	}
	if p.inlineActive {
		termPrintln(p.w, "")
		p.inlineActive = false
	}
	p.PhaseInline(transferProgressFinishingFmt)
}

// SFTPProgress reports copy percentage during SFTP transfer.
func (p *TransferProgress) SFTPProgress(written, total int64) {
	if p == nil {
		return
	}
	p.bytes = written
	if total <= 0 {
		return
	}
	pct := written * 100 / total
	if pct == p.lastPct && written != total {
		return
	}
	p.reportPercent(pct)
}

// Bytes returns the last known transferred byte count.
func (p *TransferProgress) Bytes() int64 {
	if p == nil {
		return 0
	}
	return p.bytes
}

// SetBytes records the transferred byte count for done messages.
func (p *TransferProgress) SetBytes(n int64) {
	if p != nil {
		p.bytes = n
	}
}

func (p *TransferProgress) inlineTTYEnabled() bool {
	if transferProgressInlineTTY != nil {
		return transferProgressInlineTTY(p)
	}
	if p.w != os.Stderr {
		return false
	}
	return xterm.IsTerminal(int(os.Stderr.Fd()))
}

func formatTransferStart(verb, src, dst string, size int64) string {
	if size >= 0 {
		return fmt.Sprintf("%s %s → %s (%d bytes)…", verb, src, dst, size)
	}
	return fmt.Sprintf("%s %s → %s…", verb, src, dst)
}
