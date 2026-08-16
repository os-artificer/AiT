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

package shared

// SectionFlagsJSONWatch is the Flags help body for metric one-shots (cpu/mem/disk/net).
const SectionFlagsJSONWatch = "" +
	"      --json               emit JSON\n" +
	"      --watch              refresh until interrupted\n" +
	"      --interval duration  sample interval for --watch (default 1s)\n"

// SectionFlagsProc is the Flags help body for proc (JSON/watch plus limit/sort).
const SectionFlagsProc = SectionFlagsJSONWatch +
	"      --limit int          max process rows (default 10)\n" +
	"      --sort string        primary sort: cpu|mem|net|disk (default cpu)\n"

// SectionFlagsTop is the Flags help body for the live dashboard.
const SectionFlagsTop = "" +
	"      --interval duration  refresh interval (default 1s)\n" +
	"      --limit int          optional PROC sample cap (default: all processes)\n" +
	"      --sort string        PROC primary sort: cpu|mem|net|disk (default cpu)\n"
