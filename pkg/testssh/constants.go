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

const (
	// DefaultUser is the password-auth username for the mock sshd.
	DefaultUser = "dev"
	// DefaultPassword is the password-auth secret for the mock sshd (tests only).
	DefaultPassword = "devpass"
	// DefaultHostname is returned by the mock `hostname` command.
	DefaultHostname = "mockhost.example.com"
	// DefaultAddr binds loopback with an ephemeral port.
	DefaultAddr = "127.0.0.1:0"
	// WelcomeName is the seeded file under the mock filesystem root.
	WelcomeName = "welcome.txt"
	// WelcomeBody is the contents of WelcomeName.
	WelcomeBody = "hello from mock ssh\n"
	// ProbeName is a second seeded file used by get/put roundtrips.
	ProbeName = "probe.txt"
	// ProbeBody matches the Python e2e transfer probe marker.
	ProbeBody           = "ait-e2e-probe\n"
	identityFileName    = "id_ed25519"
	promptReadySentinel = ".__AIT_PROMPT_READY__"
)
