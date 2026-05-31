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

package lrzsz

import "github.com/os-artificer/AiT/internal/pkg/errhint"

// ErrMissing is returned when the remote active shell lacks sz/rz (lrzsz).
var ErrMissing = errhint.NewSentinel("remote shell missing lrzsz (sz/rz)")

// ErrInstallDeclined is returned when the user declines automatic lrzsz installation.
var ErrInstallDeclined = errhint.NewSentinel("lrzsz install declined")

// ErrInstallFailed is returned when automatic lrzsz installation did not succeed.
var ErrInstallFailed = errhint.NewSentinel("lrzsz install failed")

// ErrNonInteractiveInstall is returned when auto-install needs a TTY confirm.
var ErrNonInteractiveInstall = errhint.NewSentinel("lrzsz install requires interactive confirmation")
