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
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
	success bool
	method  string
	commits []client.PRCommit
	labels  []string
}

func (m *MockClient) CreatePRComment(org, repo, number, comment string) bool {
	m.method = "CreatePRComment"
	return m.success
}

func (m *MockClient) AddPRLabels(org, repo, number string, labels []string) bool {
	m.method = "AddPRLabels"
	return m.success
}

func (m *MockClient) RemovePRLabels(org, repo, number string, labels []string) bool {
	m.method = "RemovePRLabels"
	return m.success
}

func (m *MockClient) CheckIfPRCreateEvent(evt *client.GenericEvent) bool {
	m.method = "CheckIfPRCreateEvent"
	return m.success
}

func (m *MockClient) CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) bool {
	m.method = "CheckIfPRSourceCodeUpdateEvent"
	return m.success
}

func (m *MockClient) GetPullRequestCommits(org, repo, number string) ([]client.PRCommit, bool) {
	m.method = "GetPullRequestCommits"
	return m.commits, m.success
}

func (m *MockClient) GetPullRequestLabels(org, repo, number string) ([]string, bool) {
	m.method = "GetPullRequestLabels"
	return m.labels, m.success
}

func (m *MockClient) Value() (success bool, method string, commits []client.PRCommit, labels []string) {
	return m.success, m.method, m.commits, m.labels
}

func (m *MockClient) Set(success bool, method string, commits []client.PRCommit, labels []string) {
	m.success = success
	m.method = method
	m.commits = commits
	m.labels = labels
}

func TestRemoveLabels(t *testing.T) {

	mockClient := new(MockClient)
	bot := &robot{cli: mockClient, cnf: &configuration{
		CommentUpdateLabelFailed: "%s, 1123, %s",
	}}

	cli, ok := bot.cli.(*MockClient)
	assert.Equal(t, true, ok)
	case1 := "No labels to remove"
	cli.method = case1
	// No labels to remove
	bot.removeLabels("org", "repo", "number", "commenter", []string{})
	assert.Equal(t, case1, cli.method)

	case2 := "RemovePRLabels"
	cli.method = case2
	cli.success = true
	// Successfully remove labels
	bot.removeLabels("org", "repo", "number", "commenter", []string{"label1"})
	assert.Equal(t, case2, cli.method)

	case3 := "CreatePRComment"
	cli.method = case3
	cli.success = false
	// Failed to remove labels
	bot.removeLabels("org", "repo", "number", "commenter", []string{"label1"})
	assert.Equal(t, case3, cli.method)

}
