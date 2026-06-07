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
	"strings"
	"time"
)

const (
	zmodemStripCANB0     = "**\x18B0"
	zmodemStripCANB00    = "\x18B00"
	zmodemStripCANB01    = "\x18B01"
	zmodemStripCANB02    = "\x18B02"
	zmodemStripZACK      = "\x18\x11"
	zmodemStripRzWaiting = "rz waiting"
	zmodemStripSzWaiting = "sz waiting"
)

// zmodemStripFragments are scrollback substrings removed after ZMODEM transfer.
var zmodemStripFragments = []string{
	zmodemStripCANB0,
	zmodemStripCANB00,
	zmodemStripCANB01,
	zmodemStripCANB02,
	zmodemStripZACK,
	zmodemStripRzWaiting,
	zmodemStripSzWaiting,
}

// zmodemDrainWaitHook waits for PTY quiet during ZMODEM tail drain (overridable in tests).
var zmodemDrainWaitHook = func(c *Client, drainAt time.Time) {
	c.waitRemoteOutputQuietSince(drainAt, zmodemTailQuietPeriod(c), remoteOutputQuietMaxWait)
}

func zmodemTailQuietPeriod(c *Client) time.Duration {
	if c != nil && c.ChainMode() {
		return transferChainCaptureQuiet
	}
	return remoteOutputQuietPeriod
}

// stripZModemScrollback removes known ZMODEM protocol fragments and short control runs from scrollback.
func (c *Client) stripZModemScrollback() {
	if c == nil || c.scroll == nil {
		return
	}

	c.scroll.StripSubstrings(zmodemStripFragments...)
	c.scroll.ReplaceWith(stripZModemControlRunsString)
}

func stripZModemControlRunsString(s string) string {
	if s == "" {
		return s
	}

	var out strings.Builder
	out.Grow(len(s))

	var ctrlRun strings.Builder

	flushCtrl := func() {
		if ctrlRun.Len() >= 2 {
			ctrlRun.Reset()
			return
		}
		out.WriteString(ctrlRun.String())
		ctrlRun.Reset()
	}

	for _, r := range s {
		if isZModemControlRune(r) {
			ctrlRun.WriteRune(r)
			continue
		}
		flushCtrl()
		out.WriteRune(r)
	}
	flushCtrl()
	return out.String()
}

func isZModemControlRune(r rune) bool {
	if r == '\t' || r == '\n' || r == '\r' {
		return false
	}
	return r <= 0x1f || r == 0x7f
}

// drainZModemPostWait absorbs late PTY bytes into the ZMODEM engine while it is still active.
func (c *Client) drainZModemPostWait() {
	if c == nil || !c.zmodemActive.Load() {
		return
	}

	zmodemDrainWaitHook(c, time.Now())
}

// drainZModemTail waits for PTY quiet while stdout is suppressed, then strips protocol noise from scrollback.
func (c *Client) drainZModemTail() {
	if c == nil {
		return
	}

	zmodemDrainWaitHook(c, time.Now())
	c.stripZModemScrollback()
}
