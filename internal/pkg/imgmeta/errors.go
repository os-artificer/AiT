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

package imgmeta

import "errors"

var (
	// ErrUnsupported is returned when the file is not a recognized image container.
	ErrUnsupported = errors.New("unsupported image format")
	// ErrEmptyPath is returned when Extract is called with an empty path.
	ErrEmptyPath = errors.New("image path required")
	// ErrPartialFailure reports that one or more input files failed while others may have succeeded.
	ErrPartialFailure = errors.New("one or more files failed")
	// ErrNothingToChange is returned when Patch/Clear is called with an empty request.
	ErrNothingToChange = errors.New("no metadata fields specified")
)
