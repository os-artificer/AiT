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

// Package route selects get/put backends (ZMODEM vs SFTP) for in-session transfer.
package route

// Backend is the file transfer implementation for a planned transfer.
type Backend int

const (
	// BackendZModem uses lrzsz-compatible ZMODEM over the interactive PTY.
	BackendZModem Backend = iota
	// BackendSFTP uses SFTP over the dial SSH connection.
	BackendSFTP
)

// Plan describes how an in-session get/put should run.
type Plan struct {
	Backend   Backend
	ChainOnly bool
	AutoChain bool
	LrzszOK   bool
}
