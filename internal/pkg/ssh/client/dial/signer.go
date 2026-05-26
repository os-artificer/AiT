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

package dial

import (
	"os"

	"golang.org/x/crypto/ssh"
)

// LoadSigner reads path and parses an OpenSSH private key.
// passphrase is tried only when the key parses as encrypted; passing an SSH
// login password by mistake never decrypts an unencrypted key.
// Returns ssh.ParsePrivateKey or ssh.ParsePrivateKeyWithPassphrase errors.
func LoadSigner(path, passphrase string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(b)
	if err == nil {
		return signer, nil
	}

	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(b, []byte(passphrase))
	}
	return nil, err
}
