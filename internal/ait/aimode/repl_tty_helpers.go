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

package aimode

import (
	"bufio"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/shell"
)

func (o Options) withTTYConfirm(br *bufio.Reader) Options {
	if o.ConfirmRead != nil {
		return o
	}

	out := o.Out
	if out == nil {
		out = os.Stdout
	}

	o.ConfirmRead = func(prompt string) (ConfirmResult, error) {
		aimWrite(out, prompt)
		line, err := shell.ReadTTYLineNoHistory(br, out, "")
		if err != nil {
			return ConfirmResult{}, err
		}
		allowPattern := true
		allowSession := true
		if strings.Contains(prompt, "HIGH-RISK") {
			allowPattern = false
			allowSession = false
		}
		return parseConfirmChoice(line, allowPattern, allowSession), nil
	}
	return o
}
