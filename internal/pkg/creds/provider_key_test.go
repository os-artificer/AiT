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
	"testing"

	"github.com/99designs/keyring"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

func testKeyringEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
}

func TestProviderAPIKey_AliasMigrationM9(t *testing.T) {
	testKeyringEnv(t)

	const key = "secret-deep"
	if err := SetDeepseekAPIKey(key); err != nil {
		t.Fatal(err)
	}

	got, err := GetProviderAPIKey(config.ProviderDeepseek)
	if err != nil || got != key {
		t.Fatalf("ai path get=%q err=%v", got, err)
	}
	if alias, err := GetDeepseekAPIKey(); err != nil || alias != key {
		t.Fatalf("alias get=%q err=%v", alias, err)
	}

	kr, err := openRing()
	if err != nil {
		t.Fatal(err)
	}
	item, err := kr.Get(providerAPIKeyKey(config.ProviderDeepseek))
	if err != nil || string(item.Data) != key {
		t.Fatalf("keyring ai/deepseek/api_key: err=%v data=%q", err, item.Data)
	}
}

func TestProviderAPIKey_LegacyDeepseekMigrates(t *testing.T) {
	testKeyringEnv(t)

	const key = "legacy-key"
	kr, err := openRing()
	if err != nil {
		t.Fatal(err)
	}
	if err := kr.Set(keyring.Item{Key: "deepseek/api_key", Data: []byte(key)}); err != nil {
		t.Fatal(err)
	}

	got, err := GetProviderAPIKey(config.ProviderDeepseek)
	if err != nil || got != key {
		t.Fatalf("got=%q err=%v", got, err)
	}

	item, err := kr.Get(providerAPIKeyKey(config.ProviderDeepseek))
	if err != nil || string(item.Data) != key {
		t.Fatalf("migrated key missing: err=%v", err)
	}
}

func TestProviderAPIKey_MinimaxAlias(t *testing.T) {
	testKeyringEnv(t)

	const key = "mini-secret"
	if err := SetMinimaxAPIKey(key); err != nil {
		t.Fatal(err)
	}
	got, err := GetProviderAPIKey(config.ProviderMinimax)
	if err != nil || got != key {
		t.Fatalf("got=%q err=%v", got, err)
	}
}
