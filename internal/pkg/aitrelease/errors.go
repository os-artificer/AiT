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

package aitrelease

import "errors"

// ErrNoRelease indicates GitHub has no published release yet.
var ErrNoRelease = errors.New("no GitHub release found")

// ErrNoAsset indicates no release asset matches this platform.
var ErrNoAsset = errors.New("no release asset for this platform")

// ErrInstallRoot indicates the install directory could not be resolved or is not writable.
var ErrInstallRoot = errors.New("install root unavailable or not writable")

// ErrNothingToUpdate indicates update had no installed official plugins to refresh.
var ErrNothingToUpdate = errors.New("nothing to update")

// ErrBundleInvalid indicates the release archive failed validation or extraction.
var ErrBundleInvalid = errors.New("invalid release bundle")

// ErrUnknownPlugin indicates a requested name is not an official plugin.
var ErrUnknownPlugin = errors.New("unknown official plugin")

// ErrUpgradeExtraArgs indicates upgrade was called with trailing arguments.
var ErrUpgradeExtraArgs = errors.New("upgrade accepts no arguments")
