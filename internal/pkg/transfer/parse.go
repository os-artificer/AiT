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

package transfer

import (
	"fmt"
	"strings"
)

// ParseRemoteDest parses user@host:/path and user@[ipv6]:/path forms.
// Returns an error when host or path cannot be determined.
func ParseRemoteDest(dst string) (user, host, path string, err error) {
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
			return "", "", "", fmt.Errorf(errInvalidRemotePathFmt, dst)
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
		return "", "", "", fmt.Errorf(errInvalidRemotePathFmt, dst)
	}
	return user, host, path, nil
}
