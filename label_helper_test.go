// Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestMatchLabelFromCommentLine(t *testing.T) {
	type args struct {
		commandLine string
		reg         *regexp.Regexp
	}
	testCases := []struct {
		desc string
		in   args
		out  string
	}{
		{
			"the command is missing space",
			args{
				"/kindbug",
				regexpCommentByAnyoneToAddLabel,
			},
			"",
		},
		{
			"there is a space after the slash",
			args{
				"/ kind bug",
				regexpCommentByAnyoneToAddLabel,
			},
			"",
		},
		{
			"there are some characters that are not allowed",
			args{
				"/kind bug/1",
				regexpCommentByAnyoneToAddLabel,
			},
			"",
		},
		{
			"a correct command line to add 'kind' label, it contains a space",
			args{
				"/kind bug",
				regexpCommentByAnyoneToAddLabel,
			},
			"kind/bug",
		},
		{
			"a correct command line to add 'priority' label, it contains a tab",
			args{
				"/priority\thigh",
				regexpCommentByAnyoneToAddLabel,
			},
			"priority/high",
		},
		{
			"a correct command line to add 'sig' label, it mixes space and tab",
			args{
				"/sig\t Kernel",
				regexpCommentByAnyoneToAddLabel,
			},
			"sig/Kernel",
		},
		{
			"a correct command line to add 'good' label, it contains some space",
			args{
				"/good  thing",
				regexpCommentByAnyoneToAddLabel,
			},
			"good/thing",
		},
		{
			"a correct command line to add 'sig' label, it contains some tab",
			args{
				"/sig\t\tKernel",
				regexpCommentByAnyoneToAddLabel,
			},
			"sig/Kernel",
		},
		{
			"a correct command line to remove 'sig' label, it contains a space",
			args{
				"/remove-sig Community",
				regexpCommentByAnyoneRemoveLabel,
			},
			"sig/Community",
		},
		{
			"a correct command line to remove 'priority' label, it contains a tab",
			args{
				"/remove-priority\tlow",
				regexpCommentByAnyoneRemoveLabel,
			},
			"priority/low",
		},
	}
	for i := range testCases {
		t.Run(testCases[i].desc, func(t *testing.T) {
			got := matchLabelFromCommentLine(testCases[i].in.commandLine, testCases[i].in.reg)
			assert.Equal(t, testCases[i].out, got)
		})
	}
}

func TestMatchLabels(t *testing.T) {
	testCases := []struct {
		desc string
		in   string
		out  [2][]string
	}{
		{
			"a wrong command line that contains nothing",
			"",
			[2][]string{nil, nil},
		},
		{
			"a wrong command line that only contains spaces",
			"  ",
			[2][]string{nil, nil},
		}, {
			"a wrong command line that mixes spaces and newline",
			" \n \n ",
			[2][]string{nil, nil},
		},
		{
			"a wrong command line that is invalid",
			"/123 ooooogf",
			[2][]string{nil, nil},
		},
		{
			"a correct command line to add 'kind' label",
			"/kind question ",
			[2][]string{{"kind/question"}, nil},
		},
		{
			"a correct command line to add multi-'kind' label",
			"/kind question \n /kind help-wanted",
			[2][]string{{"kind/question", "kind/help-wanted"}, nil},
		},
		{
			"a correct command line to add 'kind' and 'sig' label",
			"/kind question \n /sig release-ROS",
			[2][]string{{"kind/question", "sig/release-ROS"}, nil},
		},
		{
			"a correct command line to update 'kind' label",
			"/remove-kind question \n /kind bug",
			[2][]string{{"kind/bug"}, {"kind/question"}},
		},
		{
			"a correct command line to remove 'kind' and 'priority' label",
			"/remove-kind bug \n /remove-priority low",
			[2][]string{nil, {"kind/bug", "priority/low"}},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].desc, func(t *testing.T) {
			got1, got2 := matchLabels(testCases[i].in)
			assert.Equal(t, testCases[i].out[0], got1)
			assert.Equal(t, testCases[i].out[1], got2)
		})
	}
}

func TestCheckIntersection(t *testing.T) {
	type result struct {
		b bool
		s string
	}

	testCases := []struct {
		desc string
		in   [2][]string
		out  result
	}{
		{
			"add 1 label",
			[2][]string{{"kind/task"}, nil},
			result{false, ""},
		},
		{
			"add 2 label",
			[2][]string{{"kind/bug", "priority/low"}, nil},
			result{false, ""},
		},
		{
			"remove 2 label",
			[2][]string{nil, {"kind/bug", "priority/low"}},
			result{false, ""},
		},
		{
			"add 1 label, remove 1 label",
			[2][]string{{"kind/bug"}, {"kind/task"}},
			result{false, ""},
		},
		{
			"add 1 label, remove 1 label",
			[2][]string{{"kind/task"}, {"kind/task"}},
			result{true, "kind/task"},
		},
		{
			"add 2 label, remove 2 label",
			[2][]string{{"kind/task", "kind/CVE"}, {"kind/task", "kind/cve"}},
			result{true, "kind/cve**, **kind/task"},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].desc, func(t *testing.T) {
			got1, got2 := checkIntersection(testCases[i].in[0], testCases[i].in[1])
			assert.Equal(t, testCases[i].out.b, got1)
			assert.Equal(t, testCases[i].out.s, got2)
		})
	}
}
