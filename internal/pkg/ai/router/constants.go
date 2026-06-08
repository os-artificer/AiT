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

package router

const (
	errAllProvidersFailedFmt  = "ai: all providers failed (last: %s)"
	errAllProvidersFailedBody = "ai: all providers failed; last attempt %s:"
	failoverLogFmt            = "AiT: AI failover %s -> %s (%v)\n"
	retrySubstrEmptyResponse  = "empty response"
	retrySubstrEmptyChoices   = "empty choices"

	httpStatusTooManyRequests = 429
	httpStatusInternalError   = 500
	httpStatusBadGateway      = 502
	httpStatusUnavailable     = 503
	httpStatusGatewayTimeout  = 504
)
