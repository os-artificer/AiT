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

package meta

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func metaBlockedErr(subcmd, detail string) error {
	return errhint.WithTryMsg(fmt.Sprintf("%q cannot run inside SSH", subcmd),
		detail,
		"exit",
		"!ait help",
	)
}

func metaExecutorNotConfiguredErr() error {
	return errhint.WithTryMsg("meta executor not configured",
		"Internal wiring error; restart ait.",
	)
}
