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

package session

import (
	"errors"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// wrapLoadNotFound augments ErrSessionNotFound when idOrAlias looks like a truncated list ID.
func wrapLoadNotFound(idOrAlias string, err error) error {
	if !errors.Is(err, store.ErrSessionNotFound) || !isListDisplayIDToken(idOrAlias) {
		return err
	}
	return errhint.WithTry(err,
		"The list ID column is for display only.",
		tryLoadByAlias, tryLoadByFullID, tryLoadInteractive, trySessionHelp,
	)
}

// isListDisplayIDToken reports whether s matches the 8-character hex prefix shown in session list ID column.
func isListDisplayIDToken(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
