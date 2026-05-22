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

package creds

import (
	"os"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	"github.com/99designs/keyring"
)

const defaultFileKeyringPassword = "ait-test"

func resolveFileKeyringDir() (string, error) {
	if dir := os.Getenv(envFileDir); dir != "" {
		return dir, nil
	}
	return platform.KeyringDir()
}

func fileKeyringPasswordFunc() func(string) (string, error) {
	return func(string) (string, error) {
		if p := os.Getenv(envFilePassword); p != "" {
			return p, nil
		}
		return defaultFileKeyringPassword, nil
	}
}

func applyFileKeyringConfig(cfg *keyring.Config) error {
	dir, err := resolveFileKeyringDir()
	if err != nil {
		return err
	}
	cfg.FileDir = dir
	cfg.FilePasswordFunc = fileKeyringPasswordFunc()
	return nil
}
