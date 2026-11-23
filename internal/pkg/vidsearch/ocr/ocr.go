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

// Package ocr runs tesseract over extracted frames and matches queries.
package ocr

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

var (
	rePhone = regexp.MustCompile(`(?:\+?\d[\d\-\s]{6,}\d)`)
	// Simplified CN plate-ish + alnum token (best-effort).
	rePlate = regexp.MustCompile(`[\p{Han}]?[A-Za-z0-9]{5,8}`)
)

// Index maps frame path → OCR text.
type Index struct {
	ByFrame map[string]string
	Frames  []frames.Frame
}

// Build runs tesseract on each frame (eng+chi_sim when available).
func Build(ctx context.Context, fr []frames.Frame, tessPath string) (Index, error) {
	if tessPath == "" {
		tessPath = "tesseract"
	}
	idx := Index{ByFrame: make(map[string]string, len(fr)), Frames: fr}
	for _, f := range fr {
		select {
		case <-ctx.Done():
			return idx, ctx.Err()
		default:
		}
		text, err := runTesseract(ctx, tessPath, f.Path)
		if err != nil {
			// Keep going; empty text for failed frames.
			idx.ByFrame[f.Path] = ""
			continue
		}
		idx.ByFrame[f.Path] = text
	}
	return idx, nil
}

func runTesseract(ctx context.Context, bin, image string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, image, "stdout", "-l", "eng+chi_sim", "--psm", "6")
	out, err := cmd.Output()
	if err != nil {
		// Retry eng-only.
		cmd = exec.CommandContext(ctx, bin, image, "stdout", "-l", "eng", "--psm", "6")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("tesseract: %w", err)
		}
	}
	return string(out), nil
}

// Search finds query matches in the OCR index.
func Search(idx Index, query, mode string) []hit.Hit {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	var hits []hit.Hit
	qlower := strings.ToLower(q)
	for _, f := range idx.Frames {
		text := idx.ByFrame[f.Path]
		if text == "" {
			continue
		}
		snippet, kind, score := matchText(text, q, qlower)
		if snippet == "" {
			continue
		}
		hits = append(hits, hit.Hit{
			TSec:    f.TSec,
			Frame:   f.Path,
			Kind:    kind,
			Query:   q,
			Snippet: oneLine(snippet),
			Score:   score,
			Mode:    mode,
		})
	}
	return hits
}

func matchText(text, query, qlower string) (snippet, kind string, score float64) {
	lower := strings.ToLower(text)
	if i := strings.Index(lower, qlower); i >= 0 {
		return extractSnippet(text, i, len(query)), hit.KindText, 1
	}
	// Regex helpers when query looks like plate/phone keywords or digits.
	if looksPhoneQuery(query) {
		if m := rePhone.FindString(text); m != "" {
			return m, hit.KindRegex, 0.9
		}
	}
	if looksPlateQuery(query) {
		if m := rePlate.FindString(text); m != "" && strings.Contains(strings.ToLower(m), strings.ToLower(stripHan(query))) {
			return m, hit.KindRegex, 0.85
		}
		// Also return any plate-like if query is generic "车牌"
		if isGenericPlateWord(query) {
			if m := rePlate.FindString(text); m != "" {
				return m, hit.KindRegex, 0.7
			}
		}
	}
	return "", "", 0
}

func looksPhoneQuery(q string) bool {
	ql := strings.ToLower(q)
	if strings.Contains(ql, "phone") || strings.Contains(q, "电话") || strings.Contains(q, "手机") {
		return true
	}
	digit := 0
	for _, r := range q {
		if unicode.IsDigit(r) {
			digit++
		}
	}
	return digit >= 7
}

func looksPlateQuery(q string) bool {
	ql := strings.ToLower(q)
	return strings.Contains(q, "车牌") || strings.Contains(ql, "plate") || isGenericPlateWord(q) || hasHanLetterDigit(q)
}

func isGenericPlateWord(q string) bool {
	return q == "车牌" || strings.EqualFold(q, "plate") || strings.EqualFold(q, "license")
}

func hasHanLetterDigit(q string) bool {
	hasHan, hasAlnum := false, false
	for _, r := range q {
		if unicode.Is(unicode.Han, r) {
			hasHan = true
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlnum = true
		}
	}
	return hasHan && hasAlnum
}

func stripHan(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractSnippet(text string, i, qlen int) string {
	start := i - 16
	if start < 0 {
		start = 0
	}
	end := i + qlen + 16
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}
