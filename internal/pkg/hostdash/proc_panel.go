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

package hostdash

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/charmbracelet/lipgloss"
)

var selectedRowStyle = lipgloss.NewStyle().Reverse(true)

const (
	procModeList = iota
	procModeSearch
	procModeDetail
	procModeConfirmTERM
	procModeConfirmKILL
)

const (
	procKillHintDetail = "x kill · esc back"
	procStatusSIGTERM  = "SIGTERM"
	procStatusSIGKILL  = "SIGKILL"

	keySearch = "/"
	keyEsc    = "esc"
	keyEscape = "escape"
	keyEnter  = "enter"
	keyKill   = "x"
	keyQuit   = "q"
	keyYes    = "y"
	keyNo     = "n"
	keyUp     = "up"
	keyDown   = "down"
	keyK      = "k"
	keyJ      = "j"
	keyBS     = "backspace"
	keyCtrlH  = "ctrl+h"
)

// ProcPanel shows the top process table with optional search, detail, and kill confirm.
type ProcPanel struct {
	all        []hoststats.ProcStat
	filtered   []hoststats.ProcStat
	order      []int32 // frozen PID order while search/detail/confirm; nil in list mode
	mode       int
	query      string
	cursor     int
	focusPID   int32
	killPID    int32
	killName   string
	killErr    string
	returnMode int
	statusMsg  string
	killFn     func(pid int32, force bool) error
}

// NewProcPanel returns a process list panel.
func NewProcPanel() *ProcPanel { return &ProcPanel{} }

// Title implements Panel.
func (p *ProcPanel) Title() string { return titlePROC }

// SampleKeepPIDs returns PIDs that must stay in the next host sample while search,
// detail, or kill-confirm is active so frozen rows keep refreshing outside top-N.
func (p *ProcPanel) SampleKeepPIDs() []int32 {
	if !p.orderFrozen() {
		return nil
	}
	if len(p.order) > 0 {
		out := append([]int32(nil), p.order...)
		if p.focusPID != 0 {
			out = appendKeepPID(out, p.focusPID)
		}
		return out
	}
	if p.focusPID != 0 {
		return []int32{p.focusPID}
	}
	return nil
}

// ProcessCount returns the number of processes in the current sample pool.
func (p *ProcPanel) ProcessCount() int {
	return len(p.all)
}

// Update implements Panel.
func (p *ProcPanel) Update(s hoststats.Snapshot) {
	if p.orderFrozen() {
		p.all = mergeProcsPreserveOrder(p.order, s.Processes)
		p.order = procPIDs(p.all)
	} else {
		p.all = append([]hoststats.ProcStat(nil), s.Processes...)
	}
	p.refreshFiltered()
	if p.mode == procModeDetail && !p.focusStillPresent() {
		p.mode = procModeSearch
		p.focusPID = 0
	}
}

// HandleKey handles PROC search/detail/kill keys. Returns true when the key is consumed.
func (p *ProcPanel) HandleKey(key string) bool {
	switch p.mode {
	case procModeConfirmTERM, procModeConfirmKILL:
		return p.handleConfirmKey(key)
	case procModeDetail:
		return p.handleDetailKey(key)
	case procModeSearch:
		return p.handleSearchKey(key)
	default:
		return p.handleListKey(key)
	}
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *ProcPanel) View(width, height int) string {
	cw, ch := contentBoxSize(width, height)
	var body string
	switch p.mode {
	case procModeConfirmTERM, procModeConfirmKILL:
		body = p.renderKillConfirm()
	case procModeDetail:
		body = p.renderDetail(ch)
	case procModeSearch:
		body = p.renderSearchTable(cw, ch)
	default:
		body = p.renderListTable(cw, ch)
	}
	return renderBox(strings.TrimRight(body, "\n"), width, height)
}

func (p *ProcPanel) handleListKey(key string) bool {
	if key == keySearch {
		p.enterSearch()
		return true
	}
	return false
}

func (p *ProcPanel) handleSearchKey(key string) bool {
	switch key {
	case keyEsc, keyEscape:
		p.handleSearchDismiss()
		return true
	case keyEnter:
		if len(p.filtered) == 0 {
			return true
		}
		p.focusPID = p.filtered[p.cursor].PID
		p.mode = procModeDetail
		p.statusMsg = ""
		return true
	case keyKill:
		return p.beginKillConfirm()
	case keyUp, keyK:
		p.moveCursor(-1)
		return true
	case keyDown, keyJ:
		p.moveCursor(1)
		return true
	case keyBS, keyCtrlH:
		if p.query == "" {
			return true
		}
		_, size := utf8.DecodeLastRuneInString(p.query)
		p.query = p.query[:len(p.query)-size]
		p.cursor = 0
		p.refreshFiltered()
		return true
	case keyQuit:
		p.handleSearchDismiss()
		return true
	}
	if isProcSearchRune(key) {
		p.query += key
		p.cursor = 0
		p.refreshFiltered()
		return true
	}
	return false
}

func (p *ProcPanel) handleDetailKey(key string) bool {
	switch key {
	case keyEsc, keyEscape:
		p.mode = procModeSearch
		p.focusPID = 0
		p.statusMsg = ""
		p.refreshFiltered()
		p.clampCursor()
		return true
	case keyKill:
		return p.beginKillConfirm()
	default:
		return false
	}
}

func (p *ProcPanel) handleConfirmKey(key string) bool {
	switch key {
	case keyNo, keyEsc, keyEscape:
		p.cancelKillConfirm()
		return true
	case keyYes:
		p.applyKillConfirm()
		return true
	default:
		return true // consume all keys in confirm so they do not leak to search
	}
}

func (p *ProcPanel) beginKillConfirm() bool {
	pid, name, ok := p.selectedKillTarget()
	if !ok {
		return true
	}
	p.returnMode = p.mode
	p.killPID = pid
	p.killName = name
	p.killErr = ""
	p.mode = procModeConfirmTERM
	return true
}

func (p *ProcPanel) cancelKillConfirm() {
	p.mode = p.returnMode
	p.clearKillState()
}

func (p *ProcPanel) applyKillConfirm() {
	force := p.mode == procModeConfirmKILL
	err := p.doKill(p.killPID, force)
	if err == nil {
		sig := procStatusSIGTERM
		if force {
			sig = procStatusSIGKILL
		}
		p.statusMsg = fmt.Sprintf("killed pid %d (%s)", p.killPID, sig)
		p.mode = p.returnMode
		p.clearKillState()
		return
	}

	if p.mode == procModeConfirmTERM {
		p.killErr = err.Error()
		p.mode = procModeConfirmKILL
		return
	}

	p.statusMsg = fmt.Sprintf("kill failed: %s", err.Error())
	p.mode = p.returnMode
	p.clearKillState()
}

func (p *ProcPanel) clearKillState() {
	p.killPID = 0
	p.killName = ""
	p.killErr = ""
}

func (p *ProcPanel) doKill(pid int32, force bool) error {
	fn := p.killFn
	if fn == nil {
		fn = hoststats.KillProc
	}
	return fn(pid, force)
}

func (p *ProcPanel) selectedKillTarget() (pid int32, name string, ok bool) {
	switch p.mode {
	case procModeDetail:
		focus, found := p.focusProc()
		if !found {
			return 0, "", false
		}
		return focus.PID, focus.Name, true
	case procModeSearch:
		if len(p.filtered) == 0 {
			return 0, "", false
		}
		row := p.filtered[p.cursor]
		return row.PID, row.Name, true
	default:
		return 0, "", false
	}
}

func (p *ProcPanel) handleSearchDismiss() {
	if p.query != "" {
		p.query = ""
		p.cursor = 0
		p.refreshFiltered()
		return
	}
	p.leaveSearchToList()
}

func (p *ProcPanel) enterSearch() {
	p.mode = procModeSearch
	p.query = ""
	p.cursor = 0
	p.statusMsg = ""
	p.order = procPIDs(p.all)
	p.refreshFiltered()
}

func (p *ProcPanel) leaveSearchToList() {
	p.mode = procModeList
	p.cursor = 0
	p.statusMsg = ""
	p.order = nil
}

func (p *ProcPanel) orderFrozen() bool {
	switch p.mode {
	case procModeSearch, procModeDetail, procModeConfirmTERM, procModeConfirmKILL:
		return true
	default:
		return false
	}
}

func (p *ProcPanel) refreshFiltered() {
	if p.mode == procModeList {
		p.filtered = p.all
		return
	}
	p.filtered = hoststats.FilterProcs(p.query, p.all)
	p.clampCursor()
}

func (p *ProcPanel) clampCursor() {
	if len(p.filtered) == 0 {
		p.cursor = 0
		return
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
	}
}

func (p *ProcPanel) moveCursor(delta int) {
	if len(p.filtered) == 0 {
		return
	}
	p.cursor += delta
	p.clampCursor()
}

func (p *ProcPanel) focusStillPresent() bool {
	_, ok := p.focusProc()
	return ok
}

func (p *ProcPanel) focusProc() (hoststats.ProcStat, bool) {
	for _, row := range p.all {
		if row.PID == p.focusPID {
			return row, true
		}
	}
	return hoststats.ProcStat{}, false
}

func (p *ProcPanel) renderListTable(contentW, contentH int) string {
	return p.renderProcTable("PROC (top by CPU>MEM>NET>DISK)", p.all, contentW, contentH, "", -1)
}

func (p *ProcPanel) renderSearchTable(contentW, contentH int) string {
	title := fmt.Sprintf("PROC /%s", p.query)
	return p.renderProcTable(title, p.filtered, contentW, contentH, p.statusMsg, p.cursor)
}

func (p *ProcPanel) renderProcTable(title string, rows []hoststats.ProcStat, contentW, contentH int, status string, cursor int) string {
	logW := procLogColumnWidth(contentW)
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteByte('\n')
	if status != "" {
		b.WriteString(status)
		b.WriteByte('\n')
	}
	b.WriteString(hoststats.FormatProcTableHeader(logW))
	b.WriteByte('\n')
	maxRows := procRowCapacity(contentH)
	if status != "" && maxRows > 0 {
		maxRows--
	}
	for i, row := range rows {
		if i >= maxRows {
			break
		}
		line := hoststats.FormatProcRowWidth(row, logW)
		if cursor >= 0 && i == cursor {
			line = selectedRowStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *ProcPanel) renderDetail(contentH int) string {
	focus, ok := p.focusProc()
	if !ok {
		return titleStyle.Render("PROC detail") + "\n(process gone)\n"
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("PROC detail pid %d", focus.PID)))
	b.WriteByte('\n')
	if p.statusMsg != "" {
		b.WriteString(p.statusMsg)
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf("name  %s\n", focus.Name))
	cmd := focus.Cmdline
	if cmd == "" {
		cmd = "(empty)"
	}
	b.WriteString(fmt.Sprintf("cmd   %s\n", cmd))
	b.WriteString(fmt.Sprintf("cpu   %.1f%%  mem %.1f%%  disk %s\n",
		focus.CPUPercent, focus.MemPercent,
		hoststats.FormatBytes(uint64(focus.DiskBytesPerSec))))

	remain := contentH - strings.Count(b.String(), "\n") - 1
	if remain < 1 {
		remain = 1
	}
	writeCappedDetailList(&b, "peer:", hoststats.FormatProcPeerLines(focus), &remain)
	writeCappedDetailList(&b, "logs:", formatProcLogDetailLines(focus.Logs), &remain)

	anc := hoststats.ProcAncestors(p.all, focus.PID)
	ancLines := make([]string, 0, len(anc))
	for _, a := range anc {
		ancLines = append(ancLines, fmt.Sprintf("%d %s", a.PID, a.Name))
	}
	writeCappedDetailList(&b, "ancestors:", ancLines, &remain)

	kids := hoststats.ProcChildren(p.all, focus.PID)
	kidLines := make([]string, 0, len(kids))
	for _, c := range kids {
		kidLines = append(kidLines, fmt.Sprintf("%d %s  cpu %.1f%%", c.PID, c.Name, c.CPUPercent))
	}
	writeCappedDetailList(&b, "children:", kidLines, &remain)

	b.WriteString(procKillHintDetail)
	b.WriteByte('\n')
	return b.String()
}

func (p *ProcPanel) renderKillConfirm() string {
	var b strings.Builder
	name := p.killName
	if name == "" {
		name = "?"
	}
	if p.mode == procModeConfirmKILL {
		b.WriteString(titleStyle.Render("PROC kill force"))
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("SIGTERM failed for pid %d (%s)\n", p.killPID, name))
		if p.killErr != "" {
			b.WriteString(p.killErr)
			b.WriteByte('\n')
		}
		b.WriteString("Force SIGKILL? [y/N]\n")
		return b.String()
	}
	b.WriteString(titleStyle.Render("PROC kill"))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Send SIGTERM to pid %d (%s)?\n", p.killPID, name))
	b.WriteString("Confirm [y/N]\n")
	return b.String()
}
