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

import "github.com/os-artificer/AiT/internal/pkg/toolname"

var msgSuspendLocalUseSSH = "Reconnect to the session and run " + toolname.MetaLine("suspend") + " inside SSH."

// Ports and lengths.
const (
	defaultSSHPort        = 22
	listDisplayIDTokenLen = 8
)

// Errhint sentinels (errors.Is / WithTry first line — do not change wording).
const (
	errInvalidSelection                = "invalid selection"
	errEphemeralSuspendRequiresPersist = "suspend requires a saved session; use aissh --save or --alias <name>"
	errNoSavedSessions                 = "no saved sessions"
	errNoCurrentSession                = "no current session"
	errNonInteractiveTerminal          = "non-interactive terminal"
	errPickCancelled                   = "session pick cancelled"
	errCurrentSessionActiveFmt         = "current session %s is still active"
	errUnknownTransferOp               = "unknown transfer operation"
)

// Hint explanations (WithTryMsg body — unchanged).
const (
	hintInvalidSelectionExplain    = "Enter a number from the list, a full session id, or an alias."
	hintEphemeralSuspendExplain    = "Only saved sessions can be suspended to the local database."
	hintNoSavedSessionsExplain     = "Save a connection before loading or switching."
	hintNonInteractivePickExplain  = "Interactive session pick requires a TTY."
	hintPickCancelledExplain       = "Session selection was cancelled."
	hintNoCurrentSessionExplain    = "Load or switch to a session first, or suspend from inside SSH."
	hintDisplayIDOnlyExplain       = "The list ID column is for display only."
	hintActiveSessionSwitchExplain = "Suspend the current session inside SSH before switching."
)

// Try block lines (shared by load and interactive pick errors).
// Exported names live in try_lines.go; keep this block removed — use Try* constants.

// Interactive password prompts (format strings).
const (
	defaultSSHPromptUser       = "user"
	promptPasswordLoginFmt     = "Password for %s@%s: "
	promptPasswordExpiredFmt   = "Password for %s@%s (auth expired): "
	promptPassphraseExpiredFmt = "Passphrase for key %s (auth expired): "
)

// SSH auth error substring classification (isSSHAuthFailure).
const (
	sshHostKeyErrKnownHosts         = "knownhosts"
	sshHostKeyErrHostKey            = "host key"
	sshHostKeyErrHostKeyCompact     = "hostkey"
	sshHostKeyErrKeyUnknown         = "key is unknown"
	sshHostKeyErrVerificationFailed = "host key verification failed"
	sshHostKeyErrPossibleMITM       = "possible mitm"
	sshAuthErrUnableToAuthenticate  = "unable to authenticate"
	sshAuthErrPermissionDenied      = "permission denied"
	sshAuthErrNoSupportedAuth       = "no supported authentication"
	sshAuthErrNoAuthMethod          = "no authentication method"
)

// List / pick / confirm UI.
const (
	msgCancelled              = "cancelled"
	msgListEmpty              = "No saved sessions."
	pickSessionPrompt         = "Select session (number, id, or alias): "
	confirmDeleteAllPrompt    = "Delete ALL sessions? [y/N] "
	confirmDeleteSessionFmt   = "Delete session %s (%s@%s)? [y/N] "
	listHeaderNormal          = "#\tID\tTARGET\tSTATUS\tTMUX\tREMAINING"
	listHeaderVerbose         = "#\tID\tFULL_ID\tTARGET\tSTATUS\tTMUX\tREMAINING"
	labelAuthRemainingExpired = "expired"
	labelTmuxDisabled         = "-"
	labelTmuxEnabled          = "tmux"
)

// Stdout/stderr notices.
const (
	msgSuspendLocalUnsupported = "Suspend from the local shell is not supported."
	msgAlreadyOnSessionFmt     = "already on session %s"
	warnSwitchWithoutSuspend   = "warning: switching without suspend; marking current session suspended"
	errTmuxEnsureFmt           = "tmux: %w (is tmux installed on remote?)"
)
