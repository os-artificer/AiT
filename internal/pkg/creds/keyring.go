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
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

// ringOpen is the indirection used by every public function in this package
// so tests (or callers that set AIT_KEYRING_BACKEND) can pin a deterministic
// backend.
var ringOpen = defaultOpen

// Environment variables used to pin the keyring backend, primarily for
// tests and headless CI environments where touching the host keychain or
// DBus daemon is undesirable.
//
//	AIT_KEYRING_BACKEND=file   forces the encrypted file backend.
//	AIT_KEYRING_DIR=<path>     directory for the file backend (default: <DataDir>/keyring).
//	AIT_KEYRING_FILE_PASSWORD  master password for the file backend (default: ait-test for dev).
const (
	serviceName                = "ait"
	legacyMixedCaseServiceName = "AiT"
	envBackend                 = "AIT_KEYRING_BACKEND"
	envFileDir                 = "AIT_KEYRING_DIR"
	envFilePassword            = "AIT_KEYRING_FILE_PASSWORD"
)

func defaultOpen() (keyring.Keyring, error) {
	return openRingWithService(serviceName)
}

func openRingWithService(name string) (keyring.Keyring, error) {
	cfg := keyring.Config{ServiceName: name}

	if err := applyFileKeyringConfig(&cfg); err != nil {
		return nil, err
	}

	if os.Getenv(envBackend) == "file" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
	}

	return keyring.Open(cfg)
}

func openLegacyRing() (keyring.Keyring, error) {
	return openRingWithService(legacyMixedCaseServiceName)
}

func getItem(key string) (keyring.Item, error) {
	kr, err := openRing()
	if err != nil {
		return keyring.Item{}, err
	}

	item, err := kr.Get(key)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		return keyring.Item{}, err
	}

	legacy, legOpenErr := openLegacyRing()
	if legOpenErr != nil {
		return keyring.Item{}, err
	}

	item, legErr := legacy.Get(key)
	if legErr != nil {
		return keyring.Item{}, err
	}

	_ = kr.Set(item)
	return item, nil
}

func openRing() (keyring.Keyring, error) { return ringOpen() }

// SetSessionUnlock stores password or key passphrase for sessionID in the OS keyring.
// Returns keyring open or Set errors.
func SetSessionUnlock(sessionID, secret string) error {
	kr, err := openRing()
	if err != nil {
		return err
	}

	return kr.Set(keyring.Item{
		Key:  sessionKey(sessionID),
		Data: []byte(secret),
	})
}

// GetSessionUnlock returns stored unlock material for sessionID.
// Returns the secret, or keyring open/Get errors (including keyring.ErrKeyNotFound).
func GetSessionUnlock(sessionID string) (string, error) {
	_, err := openRing()
	if err != nil {
		return "", err
	}

	item, err := getItem(sessionKey(sessionID))
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

// DeleteSessionUnlock removes unlock material for sessionID from the keyring.
// Returns keyring open or Remove errors.
func DeleteSessionUnlock(sessionID string) error {
	kr, err := openRing()
	if err != nil {
		return err
	}

	return kr.Remove(sessionKey(sessionID))
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf(keyFmtSessionUnlock, sessionID)
}

// SetDeepseekAPIKey stores the Deepseek API key in the OS keyring (alias for ai/deepseek/api_key).
// Returns keyring open or Set errors from the underlying provider backend.
func SetDeepseekAPIKey(key string) error {
	return SetProviderAPIKey("deepseek", key)
}

// GetDeepseekAPIKey returns the Deepseek API key from the OS keyring.
// Returns the key string, or keyring open/Get errors (including keyring.ErrKeyNotFound).
func GetDeepseekAPIKey() (string, error) {
	return GetProviderAPIKey("deepseek")
}
