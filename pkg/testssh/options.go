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

// Config configures the mock SSH+SFTP daemon.
type Config struct {
	// Addr is the listen address; empty uses DefaultAddr.
	Addr string
	// User is the accepted username; empty uses DefaultUser.
	User string
	// Password is the accepted password; empty uses DefaultPassword.
	Password string
	// RejectPassword rejects every password attempt (wrong-password scenarios).
	RejectPassword bool
	// RootDir is the SFTP/exec filesystem root; empty means the caller must set it (Listen) or Start creates a temp dir.
	RootDir string
	// Hostname is returned by the mock hostname command; empty uses DefaultHostname.
	Hostname string
	// IdentityDir, when set, receives a generated OpenSSH-compatible client private key.
	IdentityDir string
}

// Info is the JSON payload printed by cmd/testenv-mock for pytest fixtures.
type Info struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Root         string `json:"root"`
	IdentityPath string `json:"identity_path,omitempty"`
	Welcome      string `json:"welcome"`
}
