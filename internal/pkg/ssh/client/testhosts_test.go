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

package client

const (
	testSSHUser = "root"

	// Primary SSH dial target FQDN.
	testPrimaryHostFQDN = "remote.example.com"
	// Short hostname as shown in native shell PS1 (e.g. root@remote).
	testPrimaryHostShort = "remote"

	// Dev/bastion dial target FQDN.
	testDevHostFQDN = "dev.example.com"
	// Short hostname segment in dev native PS1 (e.g. root@dev.example).
	testDevHostShort = "dev.example"
	// ssh config Host alias for testDevHostFQDN.
	testDevHostAlias = "devhost"

	// Nested ssh inner host FQDN.
	testInnerHostFQDN = "inner.example.com"
	// Short hostname as shown in inner native shell PS1.
	testInnerHostShort = "inner"
)
