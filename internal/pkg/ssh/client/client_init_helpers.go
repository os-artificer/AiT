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
	"strings"
	"time"
)

const (
	remoteOutputQuietPeriod   = 250 * time.Millisecond
	remoteOutputQuietMaxWait  = 2 * time.Second
	remoteOutputQuietPollStep = 25 * time.Millisecond
	remoteAITPromptMaxWait    = 5 * time.Second
)

func (c *Client) touchRemoteOut() {
	c.lastRemoteOut.Store(time.Now().UnixNano())
}

func remoteOutAt(last int64) time.Time {
	return time.Unix(0, last)
}

// waitRemoteOutputQuietAfter blocks until remote stdout/stderr has been quiet for quiet
// after injectAt, or maxWait elapses. Ignores lastRemoteOut timestamps before injectAt.
func (c *Client) waitRemoteOutputQuietAfter(injectAt time.Time, quiet, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		last := c.lastRemoteOut.Load()
		if last == 0 {
			time.Sleep(remoteOutputQuietPollStep)
			continue
		}
		if remoteOutAt(last).Before(injectAt) {
			time.Sleep(remoteOutputQuietPollStep)
			continue
		}
		if time.Since(remoteOutAt(last)) >= quiet {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// waitForRemoteAITPrompt polls scrollback until the AiT remote prompt marker appears or maxWait elapses.
func (c *Client) waitForRemoteAITPrompt(user, host string, maxWait time.Duration) {
	marker := remoteSetupPromptMarker(user, host)
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if strings.Contains(c.scroll.String(), marker) {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// finishRemoteSetup ends setup suppression, strips leaked setup text from scrollback, refreshes the remote PS1, and unpauses display.
func (c *Client) finishRemoteSetup(stdin io.Writer, user, host string) {
	marker := remoteSetupPromptMarker(user, host)

	c.waitForRemoteAITPrompt(user, host, remoteAITPromptMaxWait)
	c.scroll.StripSubstrings(RemoteSetupStripFragments(user, host)...)
	if tail := c.setupFilter.flush(user, host); len(tail) > 0 {
		_, _ = c.scroll.Write(tail)
	}

	hasMarker := strings.Contains(c.scroll.String(), marker)
	if c.pauseRemoteOut.Load() {
		if !hasMarker {
			c.refreshRemotePrompt(stdin)
		}
		c.remotePromptShown.Store(true)
	} else if !hasMarker {
		c.refreshRemotePrompt(stdin)
		c.remotePromptShown.Store(true)
	}

	c.remoteSetupSuppress.Store(false)
	c.pauseRemoteOut.Store(false)
}
