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

package aitplugin

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/shellext"
)

// Validate checks m against Manifest v1 rules.
// On success it may set m.Order to the default when the field was zero.
// Returns a descriptive error when the manifest must be skipped.
func Validate(m *Manifest) error {
	if m == nil {
		return fmt.Errorf("nil manifest")
	}
	if m.APIVersion != apiVersionV1 {
		return fmt.Errorf("unsupported apiVersion %d", m.APIVersion)
	}

	if err := validateSafeName("name", m.Name); err != nil {
		return err
	}
	if _, reserved := reservedNames[m.Name]; reserved {
		return fmt.Errorf("name %q is reserved", m.Name)
	}
	if err := validateSafeName("binary", m.Binary); err != nil {
		return err
	}
	if strings.TrimSpace(m.Short) == "" {
		return fmt.Errorf("short is required")
	}

	if m.Order == 0 {
		m.Order = defaultOrder
	}

	for i, f := range m.Flags {
		if strings.TrimSpace(f.Name) == "" {
			return fmt.Errorf("flags[%d].name is required", i)
		}
		switch f.Type {
		case flagTypeBool, flagTypeString, flagTypeInt:
		default:
			return fmt.Errorf("flags[%d].type %q unsupported", i, f.Type)
		}
		if strings.TrimSpace(f.Help) == "" {
			return fmt.Errorf("flags[%d].help is required", i)
		}
	}
	return nil
}

// IsReservedName reports whether name is a reserved ait REPL builtin.
func IsReservedName(name string) bool {
	_, ok := reservedNames[name]
	return ok
}

func validateSafeName(field, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !shellext.IsSafeName(name) {
		return fmt.Errorf("%s %q is not a safe name", field, name)
	}
	return nil
}
