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

import "testing"

// UseFileBackend pins the process keyring to an isolated file backend for t.
// Call from tests that exercise Set/Get API keys without touching the OS login keychain.
func UseFileBackend(t testing.TB) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(envBackend, "file")
	t.Setenv(envFileDir, dir)
	t.Setenv(envFilePassword, "ait-test")
}

// SkipIfNeedsSecrets skips tests that would require host keychain or live API keys.
// Use for optional integration paths; prefer mocks or UseFileBackend for unit coverage.
func SkipIfNeedsSecrets(t testing.TB) {
	t.Helper()
	t.Skip("skipped: requires host secrets / keychain (use file backend or mocks in unit tests)")
}
