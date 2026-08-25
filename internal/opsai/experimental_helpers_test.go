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

package opsai

import (
	"testing"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

func TestSkipAIExperimentalRequire(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cmdName string
		args    []string
		help    bool
		want    bool
	}{
		{name: "ai_help_positional", cmdName: Name, args: []string{pkghelpdoc.HelpArgCmd}, want: true},
		{name: "ai_help_short", cmdName: Name, args: []string{pkghelpdoc.HelpArgH}, want: true},
		{name: "ai_help_long", cmdName: Name, args: []string{pkghelpdoc.HelpArgHelp}, want: true},
		{name: "help_flag_changed", cmdName: Name, args: nil, help: true, want: true},
		{name: "ask_help_message", cmdName: "ask", args: []string{pkghelpdoc.HelpArgCmd}, want: false},
		{name: "ask_please_help_me", cmdName: "ask", args: []string{"please", "help", "me"}, want: false},
		{name: "ask_hi", cmdName: "ask", args: []string{"hi"}, want: false},
		{name: "ask_help_flag", cmdName: "ask", args: nil, help: true, want: true},
		{name: "ask_args_help_long", cmdName: "ask", args: []string{pkghelpdoc.HelpArgHelp}, want: true},
		{name: "agent_help_message", cmdName: "agent", args: []string{pkghelpdoc.HelpArgCmd}, want: false},
		{name: "agent_please_help_me", cmdName: "agent", args: []string{"please", "help", "me"}, want: false},
		{name: "agent_help_flag", cmdName: "agent", args: nil, help: true, want: true},
		{name: "agent_args_help_long", cmdName: "agent", args: []string{pkghelpdoc.HelpArgHelp}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{Use: tc.cmdName}
			cmd.InitDefaultHelpFlag()
			if tc.help {
				if err := cmd.Flags().Set("help", "true"); err != nil {
					t.Fatal(err)
				}
			}
			got := skipAIExperimentalRequire(cmd, tc.args)
			if got != tc.want {
				t.Fatalf("skipAIExperimentalRequire(name=%q, %v, helpChanged=%v)=%v, want %v",
					tc.cmdName, tc.args, tc.help, got, tc.want)
			}
		})
	}
}
