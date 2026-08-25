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

package agentcore

// EventKind identifies agent session events (SSOT for UI subscribers).
type EventKind string

const (
	EventSessionStarted   EventKind = "SessionStarted"
	EventGoalReceived     EventKind = "GoalReceived"
	EventTokenDelta       EventKind = "TokenDelta"
	EventPlanUpdated      EventKind = "PlanUpdated"
	EventToolStarted      EventKind = "ToolStarted"
	EventToolOutput       EventKind = "ToolOutput"
	EventToolFinished     EventKind = "ToolFinished"
	EventApprovalRequired EventKind = "ApprovalRequired"
	EventApprovalResolved EventKind = "ApprovalResolved"
	EventMessageFinal     EventKind = "MessageFinal"
	EventStepLimitReached EventKind = "StepLimitReached"
	EventSessionEnded     EventKind = "SessionEnded"
	EventWarning          EventKind = "Warning"
)

// Event is one item on the agent session stream.
type Event struct {
	Kind    EventKind
	Detail  string
	Tool    string
	Output  string
	Allowed bool
}

// Emitter receives session events. Implementations must not block the loop.
type Emitter interface {
	Emit(Event)
}

// FuncEmitter adapts a function to Emitter.
type FuncEmitter func(Event)

// Emit implements Emitter.
func (f FuncEmitter) Emit(e Event) {
	if f != nil {
		f(e)
	}
}

// NopEmitter discards events.
var NopEmitter Emitter = FuncEmitter(nil)
