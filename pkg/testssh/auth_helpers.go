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

package testssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

func serverConfig(cfg Config, hostSigner, clientSigner ssh.Signer) *ssh.ServerConfig {
	sc := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if cfg.RejectPassword || conn.User() != cfg.User || string(pass) != cfg.Password {
				return nil, fmt.Errorf("permission denied")
			}
			return nil, nil
		},
	}
	if clientSigner != nil {
		want := clientSigner.PublicKey()
		sc.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() != cfg.User || ssh.FingerprintSHA256(key) != ssh.FingerprintSHA256(want) {
				return nil, fmt.Errorf("permission denied")
			}
			return nil, nil
		}
	}
	sc.AddHostKey(hostSigner)
	return sc
}
