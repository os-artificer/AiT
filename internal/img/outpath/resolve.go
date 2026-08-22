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

package outpath

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// ConfirmFunc prompts and returns whether the user chose overwrite (y).
type ConfirmFunc func(prompt string) (bool, error)

// ResolveOpts configures destination path resolution for set/clear.
type ResolveOpts struct {
	Paths     []string
	Output    string
	OutputSet bool
	Inplace   bool
	PromptOut io.Writer
	ConfirmIn io.Reader
	Confirm   ConfirmFunc
	IsTTY     *bool
}

// Resolve returns one destination path per input path.
// Callers must validate field flags before calling Resolve so empty patches do not prompt.
func Resolve(opts ResolveOpts) ([]string, error) {
	if opts.OutputSet && opts.Output == "" {
		return nil, errhint.WithTry(ErrOutputEmpty,
			"Provide a non-empty path with -o.",
			toolname.AIImg+" set --make Canon -o out.jpg photo.jpg",
			toolname.AIImg+" set --help",
		)
	}
	if opts.OutputSet && opts.Inplace {
		return nil, errhint.WithTry(ErrOutputInplaceConflict,
			"Use either -o or --inplace/-y, not both.",
			toolname.AIImg+" set --make Canon -o out.jpg photo.jpg",
			toolname.AIImg+" set --make Canon --inplace photo.jpg",
		)
	}
	if opts.OutputSet && len(opts.Paths) != 1 {
		return nil, errhint.WithTry(ErrOutputMultiFile,
			"-o requires exactly one input file.",
			toolname.AIImg+" set --make Canon -o out.jpg photo.jpg",
			toolname.AIImg+" set --make Canon --inplace a.jpg b.jpg",
		)
	}
	if opts.OutputSet {
		if err := rejectIfDir(opts.Output); err != nil {
			return nil, err
		}
		return []string{opts.Output}, nil
	}
	if opts.Inplace {
		return append([]string(nil), opts.Paths...), nil
	}

	tty := stdinIsTTY()
	if opts.IsTTY != nil {
		tty = *opts.IsTTY
	}
	overwrite := false
	if tty {
		ok, err := confirmOverwrite(opts)
		if err != nil {
			return nil, err
		}
		overwrite = ok
	}
	if overwrite {
		return append([]string(nil), opts.Paths...), nil
	}

	dsts := make([]string, len(opts.Paths))
	for i, src := range opts.Paths {
		name, err := AutoName(src)
		if err != nil {
			return nil, err
		}
		dsts[i] = name
	}
	return dsts, nil
}

// PrintResult writes a success line for set/clear: inplace verb+src, or "wrote dst".
func PrintResult(w io.Writer, inplaceVerb, src, dst string) {
	if SamePath(src, dst) {
		_, _ = fmt.Fprintf(w, "%s %s\n", inplaceVerb, src)
		return
	}
	_, _ = fmt.Fprintf(w, "wrote %s\n", dst)
}

// SamePath reports whether src and dst refer to the same path after Clean.
func SamePath(src, dst string) bool {
	return filepath.Clean(src) == filepath.Clean(dst)
}

func rejectIfDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.IsDir() {
		return errhint.WithTry(ErrOutputIsDir,
			"Choose a file path for -o, not a directory.",
			toolname.AIImg+" set --make Canon -o out.jpg photo.jpg",
			toolname.AIImg+" set --help",
		)
	}
	return nil
}

func confirmOverwrite(opts ResolveOpts) (bool, error) {
	confirm := opts.Confirm
	if confirm == nil {
		in := opts.ConfirmIn
		if in == nil {
			in = os.Stdin
		}
		out := opts.PromptOut
		if out == nil {
			out = os.Stderr
		}
		confirm = defaultConfirm(in, out)
	}
	return confirm("Overwrite original file(s)? [y/N] ")
}

func defaultConfirm(in io.Reader, out io.Writer) ConfirmFunc {
	return func(prompt string) (bool, error) {
		_, _ = fmt.Fprint(out, prompt)
		line, err := readConfirmLine(in)
		if err != nil && err != io.EOF {
			return false, err
		}
		// EOF / blank / anything other than y → N (new file).
		return strings.EqualFold(strings.TrimSpace(line), terminal.ConfirmResponseYes), nil
	}
}

func readConfirmLine(r io.Reader) (string, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return line, err
}
