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
	"io"

	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

// BeginZModemSession prepares exclusive PTY tap for ZMODEM. Caller must invoke cleanup when done.
func (c *Client) BeginZModemSession(
	receive bool, ptyOut io.Writer, localPath, remoteName string, progress zmodem.ProgressFunc,
) (*zmodem.Session, func(), error) {
	c.pauseRemoteOut.Store(true)
	c.zmodemActive.Store(true)

	if c.stdinReader != nil {
		c.stdinReader.DrainForward()
	}

	var (
		sess *zmodem.Session
		err  error
	)
	if receive {
		sess, err = zmodem.NewReceiveSession(ptyOut, localPath, progress)
	} else {
		sess, err = zmodem.NewSendSession(ptyOut, localPath, remoteName, progress)
	}
	if err != nil {
		c.endZModemSession(sess)
		return nil, func() {}, err
	}

	c.zmodemSession = sess

	cleanup := func() {
		c.endZModemSession(sess)
	}
	return sess, cleanup, nil
}

func (c *Client) endZModemSession(sess *zmodem.Session) {
	if sess != nil {
		sess.Abort()
	}
	c.zmodemActive.Store(false)
	c.zmodemSession = nil
	c.stripZModemScrollback()
}

// feedZModemOutput routes remote PTY bytes into the active ZMODEM session.
func (c *Client) feedZModemOutput(data []byte) {
	if !c.zmodemActive.Load() {
		return
	}
	sess := c.zmodemSession
	if sess == nil {
		return
	}
	_, _ = sess.Feed(data)
}
