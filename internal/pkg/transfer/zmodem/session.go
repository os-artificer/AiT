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

// Package zmodem wraps the vendored xiwh ZMODEM stack for lrzsz-compatible PTY file transfer.
package zmodem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	xzmodem "github.com/os-artificer/AiT/pkg/third-party/xiwh/zmodem/zmodem"
)

const (
	waitPollInterval         = 50 * time.Millisecond
	handshakeTimeout         = 30 * time.Second
	teardownTimeout          = 15 * time.Second
	teardownTimeoutChain     = 30 * time.Second
	sendTeardownForceRelease = 2 * time.Second
	waitMaxDuration          = 30 * time.Minute
)

// sendTeardownForceReleaseDuration is how long send teardown waits for remote ZFIN before force release (tests).
var sendTeardownForceReleaseDuration = sendTeardownForceRelease

// handshakeTimeoutDuration is the wait for the remote sz/rz peer to start (overridable in tests).
var handshakeTimeoutDuration = handshakeTimeout

// teardownTimeoutDuration is the wait for ZFIN after the data plane completes (overridable in tests).
var teardownTimeoutDuration = teardownTimeout

// teardownTimeoutChainDuration is the chain-mode teardown limit (overridable in tests).
var teardownTimeoutChainDuration = teardownTimeoutChain

// ErrHandshakeTimeout is returned when the remote sz/rz peer does not start ZMODEM in time.
var ErrHandshakeTimeout = errors.New("zmodem: handshake timed out waiting for remote sz/rz")

// ErrNoProtocolResponse is returned when no ZMODEM frames arrive before handshake timeout.
var ErrNoProtocolResponse = errors.New("zmodem: no protocol response from remote peer")

// ErrTeardownTimeout is returned when ZFIN/idle is not reached soon after the file copy completes.
var ErrTeardownTimeout = errors.New("zmodem: teardown timed out waiting for ZFIN")

// ProgressFunc reports bytes transferred during a ZMODEM session.
type ProgressFunc func(written, total int64)

// Session runs ZMODEM as the local lrz/sz peer over a PTY byte stream.
type Session struct {
	engine  *xzmodem.ZModem
	ptyOut  io.Writer
	done    chan error
	once    sync.Once
	started atomic.Bool
	isSend  bool
	// testBlockIdle forces non-idle status after data done (tests in this package only).
	testBlockIdle atomic.Bool
	chainTeardown atomic.Bool
}

// SetChainTeardown selects the longer teardown window for multi-hop PTY transfers.
func (s *Session) SetChainTeardown(chain bool) {
	s.chainTeardown.Store(chain)
}

func (s *Session) teardownLimit() time.Duration {
	if s.chainTeardown.Load() {
		return teardownTimeoutChainDuration
	}
	return teardownTimeoutDuration
}

// NewReceiveSession configures local ZMODEM receive (lrz) after remote sz starts.
func NewReceiveSession(ptyOut io.Writer, localPath string, progress ProgressFunc) (*Session, error) {
	localAbs, err := filepath.Abs(localPath)
	if err != nil {
		return nil, err
	}

	s := &Session{done: make(chan error, 1), ptyOut: ptyOut}
	s.engine = xzmodem.New(xzmodem.ZModemConsumer{
		Writer:     ptyOut,
		EchoWriter: io.Discard,
		OnUpload: func() *xzmodem.ZModemFile {
			return nil
		},
		OnUploadSkip: func(*xzmodem.ZModemFile) {},
		OnCheckDownload: func(*xzmodem.ZModemFile) {
		},
		OnDownload: func(file *xzmodem.ZModemFile, reader io.ReadCloser) error {
			defer reader.Close()

			out, err := os.OpenFile(localAbs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
			if err != nil {
				s.finish(err)
				return err
			}
			defer out.Close()

			total := int64(0)
			if file != nil && file.Size > 0 {
				total = int64(file.Size)
			}
			_, err = copyWithProgress(out, reader, total, progress)
			s.finish(err)
			return err
		},
	})
	return s, nil
}

// NewSendSession configures local ZMODEM send (sz) after remote rz starts.
func NewSendSession(ptyOut io.Writer, localPath, remoteName string, progress ProgressFunc) (*Session, error) {
	zf, err := xzmodem.NewZModemLocalFile(localPath)
	if err != nil {
		return nil, err
	}
	if remoteName != "" {
		zf.Filename = remoteName
	}

	var uploaded bool
	total := int64(zf.Size)

	s := &Session{done: make(chan error, 1), ptyOut: ptyOut, isSend: true}
	s.engine = xzmodem.New(xzmodem.ZModemConsumer{
		Writer:     ptyOut,
		EchoWriter: io.Discard,
		OnUploadProgress: func(written, total int64) {
			s.reportSendUploadProgress(progress, written, total)
		},
		OnUpload: func() *xzmodem.ZModemFile {
			if uploaded {
				return nil
			}
			uploaded = true
			if total > 0 {
				s.reportSendUploadProgress(progress, 0, total)
			}
			return zf
		},
		OnUploadSkip: func(*xzmodem.ZModemFile) {},
		OnCheckDownload: func(*xzmodem.ZModemFile) {
		},
		OnDownload: func(*xzmodem.ZModemFile, io.ReadCloser) error {
			return errors.New("unexpected download during send session")
		},
	})
	return s, nil
}

// Feed passes remote PTY output into the protocol parser.
func (s *Session) Feed(data []byte) (int, error) {
	n, err := s.engine.Write(data)
	if s.engine.GetStatus() != xzmodem.StatusIdle {
		s.started.Store(true)
	}
	return n, err
}

// Abort sends the ZMODEM cancel sequence to the remote peer.
func (s *Session) Abort() {
	if s.ptyOut != nil {
		_, _ = s.ptyOut.Write(xzmodem.SES_ABORT_SEQ)
	}
}

// Close is an alias for Abort.
func (s *Session) Close() {
	s.Abort()
}

// EnterSendStateForTest marks the session as started in send mode (tests only).
func (s *Session) EnterSendStateForTest() {
	s.engine.EnterSendForTest()
	s.started.Store(true)
}

// ReportSendUploadProgressForTest drives send upload progress (tests only).
func (s *Session) ReportSendUploadProgressForTest(progress ProgressFunc, written, total int64) {
	s.reportSendUploadProgress(progress, written, total)
}

// Wait blocks until the transfer completes, ctx is cancelled, or the session times out.
func (s *Session) Wait(ctx context.Context) error {
	start := time.Now()
	handshakeDeadline := start.Add(handshakeTimeoutDuration)
	overallDeadline := start.Add(waitMaxDuration)
	ticker := time.NewTicker(waitPollInterval)
	defer ticker.Stop()

	dataDone := false
	var teardownDeadline time.Time
	var teardownStarted time.Time

	for {
		if err := ctx.Err(); err != nil {
			s.Abort()
			return err
		}

		status := s.engine.GetStatus()
		if status != xzmodem.StatusIdle {
			s.started.Store(true)
		}

		if dataDone && status == xzmodem.StatusIdle && !s.testBlockIdle.Load() {
			return nil
		}

		if !dataDone {
			select {
			case <-ctx.Done():
				s.Abort()
				return ctx.Err()
			case err := <-s.done:
				if err != nil {
					s.Abort()
					return err
				}
				dataDone = true
				teardownDeadline = time.Now().Add(s.teardownLimit())
				teardownStarted = time.Now()
			case <-ticker.C:
				now := time.Now()
				if !s.started.Load() && now.After(handshakeDeadline) {
					s.Abort()
					return fmt.Errorf("%w", ErrHandshakeTimeout)
				}
				if now.After(overallDeadline) {
					s.Abort()
					if !s.started.Load() {
						return fmt.Errorf("%w", ErrNoProtocolResponse)
					}
					return fmt.Errorf("zmodem: transfer timed out")
				}
			}
			continue
		}

		select {
		case <-ctx.Done():
			s.Abort()
			return ctx.Err()
		case err := <-s.done:
			if err != nil {
				s.Abort()
				return err
			}
		case <-ticker.C:
			if s.engine.HasUnread() {
				s.engine.PumpUnread()
			}
			if s.isSend && s.engine.SendFileEOF() {
				if s.engine.TryCompleteSendTeardown() {
					continue
				}
				if !teardownStarted.IsZero() &&
					time.Since(teardownStarted) >= sendTeardownForceReleaseDuration {
					s.engine.ForceRelease()
					continue
				}
			}
			if time.Now().After(teardownDeadline) {
				s.Abort()
				return fmt.Errorf("%w", ErrTeardownTimeout)
			}
			if time.Now().After(overallDeadline) {
				s.Abort()
				return fmt.Errorf("zmodem: transfer timed out")
			}
		}
	}
}

func (s *Session) finish(err error) {
	s.once.Do(func() {
		s.done <- err
	})
}

func (s *Session) reportSendUploadProgress(progress ProgressFunc, written, total int64) {
	if progress != nil {
		progress(written, total)
	}
	if total >= 0 && written >= total {
		s.finish(nil)
	}
}
