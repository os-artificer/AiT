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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// PickSession lists saved sessions on stdout and reads a selection from stdin.
// Returns the chosen session UUID, fmt.Errorf for empty/invalid input or no sessions, or store list errors.
func (m *Manager) PickSession() (string, error) {
	return m.PickSessionFrom(os.Stdin, os.Stdout)
}

// PickSessionFrom is like PickSession but reads the selection from in and writes prompts to out.
// Returns the chosen session UUID, fmt.Errorf for empty/invalid input or no sessions, or store list errors.
func (m *Manager) PickSessionFrom(in io.Reader, out io.Writer) (string, error) {
	list, err := m.Store.ListSessions()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", errhint.WithTryMsg("no saved sessions",
			"Save a connection before loading or switching.",
			tryLoginSave, tryLoginAlias, trySessionHelp,
		)
	}

	for i, s := range list {
		fmt.Fprintf(out, "  %d) %s  %s@%s:%d  [%s]\n", i+1, s.AliasOrID(), s.Username, s.Host, s.Port, s.Status)
	}

	fmt.Fprint(out, "Select session (number, id, or alias): ")
	line, err := readPickLine(in, out)
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", invalidSelectionErr()
	}

	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err == nil {
		if n < 1 || n > len(list) {
			return "", invalidSelectionErr()
		}
		return list[n-1].ID, nil
	}
	if sess, err := m.Store.GetSession(line); err == nil {
		return sess.ID, nil
	}
	return "", invalidSelectionErr()
}

// Delete removes one session after optional confirmation (unless force is true).
// Clears keyring unlock material and scrollback sidecar files. Returns store lookup/delete errors.
func (m *Manager) Delete(idOrAlias string, force bool) error {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return err
	}

	if !force && !confirmDelete(sess) {
		return nil
	}

	m.purgeSessionArtifacts(sess.ID)
	current, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if current == sess.ID {
		_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	}

	return m.Store.DeleteSession(sess.ID)
}

// DeleteAll removes every session after optional confirmation (unless force is true).
// Clears per-session keyring entries and scrollback files. Returns store.DeleteAllSessions errors.
func (m *Manager) DeleteAll(force bool) error {
	if !force {
		fmt.Print("Delete ALL sessions? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(line)) != "y" {
			fmt.Println(msgCancelled)
			return nil
		}
	}

	list, _ := m.Store.ListSessions()
	for _, s := range list {
		m.purgeSessionArtifacts(s.ID)
	}

	_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	return m.Store.DeleteAllSessions()
}

// List writes a tab-separated table of saved sessions to w (or a short message when empty).
// When verbose is true, a FULL_ID column with the complete session UUID is included.
// Returns store.ListSessions or tabwriter flush errors.
func (m *Manager) List(w io.Writer, verbose bool) error {
	list, err := m.Store.ListSessions()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Fprintln(w, "No saved sessions.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if verbose {
		fmt.Fprintln(tw, "#\tID\tFULL_ID\tTARGET\tSTATUS\tTMUX\tREMAINING")
	} else {
		fmt.Fprintln(tw, "#\tID\tTARGET\tSTATUS\tTMUX\tREMAINING")
	}
	now := time.Now().Unix()
	for i, s := range list {
		remain := time.Duration(s.AuthExpiresAt-now) * time.Second
		remainingLabel := remain.Round(time.Hour).String()
		if remain <= 0 {
			remainingLabel = "expired"
		}
		tmuxCol := "-"
		if s.TmuxEnabled {
			tmuxCol = "tmux"
		}
		target := fmt.Sprintf("%s@%s:%d", s.Username, s.Host, s.Port)
		if verbose {
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
				i+1, s.AliasOrID(), s.ID, target, s.Status, tmuxCol, remainingLabel)
		} else {
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
				i+1, s.AliasOrID(), target, s.Status, tmuxCol, remainingLabel)
		}
	}
	return tw.Flush()
}

// confirmDelete prompts for y/N confirmation and reports whether the user confirmed.
func confirmDelete(sess *store.Session) bool {
	fmt.Printf("Delete session %s (%s@%s)? [y/N] ", sess.AliasOrID(), sess.Username, sess.Host)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(line)) != "y" {
		fmt.Println(msgCancelled)
		return false
	}
	return true
}
