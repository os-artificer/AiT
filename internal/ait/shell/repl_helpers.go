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

package shell

import (
	"bufio"
	"io"
	"os"
)

// ttyReadResult is the outcome of processing one byte in readTTYLine.
type ttyReadResult struct {
	line string
	done bool
	err  error
}

func ttyHandleEOF(err error, line []byte) ttyReadResult {
	if err == io.EOF && len(line) > 0 {
		return ttyReadResult{line: string(line), done: true}
	}
	return ttyReadResult{err: err, done: true}
}

func ttyHandleCR(br *bufio.Reader, line []byte) ttyReadResult {
	if len(line) == 0 {
		return ttyReadResult{}
	}
	_, _ = os.Stdout.WriteString("\r\n")

	if br.Buffered() > 0 {
		next, err := br.ReadByte()
		if err == nil && next != '\n' {
			_ = br.UnreadByte()
		}
	}
	return ttyReadResult{line: string(line), done: true}
}

func ttyHandleBackspace(line []byte) (ttyReadResult, []byte) {
	if len(line) == 0 {
		return ttyReadResult{}, line
	}
	line = line[:len(line)-1]

	_, _ = os.Stdout.WriteString("\b \b")
	return ttyReadResult{}, line
}

func ttyHandlePrintable(b byte, line []byte) (ttyReadResult, []byte) {
	if b < 32 || b >= 127 {
		return ttyReadResult{}, line
	}
	line = append(line, b)

	_, _ = os.Stdout.Write([]byte{b})
	return ttyReadResult{}, line
}
