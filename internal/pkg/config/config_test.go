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

package config

import (
	"testing"
	"time"
)

func TestAuthTTL_DefaultWhenEmpty(t *testing.T) {
	cfg := &Config{}
	d, err := cfg.AuthTTL()
	if err != nil {
		t.Fatalf("AuthTTL: %v", err)
	}
	if d != DefaultAuthTTL {
		t.Fatalf("got %v want %v", d, DefaultAuthTTL)
	}
}

func TestAuthTTL_ClampLow(t *testing.T) {
	cfg := &Config{}
	cfg.Session.AuthTTL = "30s"
	d, err := cfg.AuthTTL()
	if err != nil {
		t.Fatalf("AuthTTL: %v", err)
	}
	if d != MinAuthTTL {
		t.Fatalf("got %v want %v", d, MinAuthTTL)
	}
}

func TestAuthTTL_ClampHigh(t *testing.T) {
	cfg := &Config{}
	cfg.Session.AuthTTL = "9999h"
	d, err := cfg.AuthTTL()
	if err != nil {
		t.Fatalf("AuthTTL: %v", err)
	}
	if d != MaxAuthTTL {
		t.Fatalf("got %v want %v", d, MaxAuthTTL)
	}
}

func TestAuthTTL_Valid(t *testing.T) {
	cfg := &Config{}
	cfg.Session.AuthTTL = "48h"
	d, err := cfg.AuthTTL()
	if err != nil {
		t.Fatalf("AuthTTL: %v", err)
	}
	if d != 48*time.Hour {
		t.Fatalf("got %v want 48h", d)
	}
}

func TestSetGetKey(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("deepseek.model", "deepseek-coder"); err != nil {
		t.Fatalf("SetKey: %v", err)
	}
	v, err := cfg.GetKey("deepseek.model")
	if err != nil || v != "deepseek-coder" {
		t.Fatalf("GetKey: v=%q err=%v", v, err)
	}
	if err := cfg.SetKey("session.auth_ttl", "not-a-duration"); err == nil {
		t.Fatal("SetKey should reject invalid duration")
	}
	if err := cfg.SetKey("unknown.key", "x"); err == nil {
		t.Fatal("SetKey should reject unknown key")
	}
}

func TestValidateDeepseekBaseURL(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("deepseek.base_url", "https://api.deepseek.com"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetKey("deepseek.base_url", "http://evil.example"); err == nil {
		t.Fatal("expected reject non-https")
	}
	if err := cfg.SetKey("deepseek.base_url", "https://evil.example"); err == nil {
		t.Fatal("expected reject non-allowlisted host")
	}
}
