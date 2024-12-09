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

type mockClient struct {
	mock.Mock
	successfulCreatePRComment                bool
	successfulAddPRLabels                    bool
	successfulAddIssueLabels                 bool
	successfulRemoveIssueLabels              bool
	successfulRemovePRLabels                 bool
	successfulCheckIfPRCreateEvent           bool
	successfulCheckIfPRSourceCodeUpdateEvent bool
	successfulGetPullRequestCommits          bool
	successfulGetPullRequestLabels           bool
	successfulGetIssueLabels                 bool
	successfulGetRepoIssueLabels             bool
	successfulCreateIssueComment             bool
	method                                   string
	commits                                  []client.PRCommit
	labels                                   []string
}

func (m *mockClient) CreatePRComment(org, repo, number, comment string) bool {
	m.method = "CreatePRComment"
	return m.successfulCreatePRComment
}

func (m *mockClient) CreateIssueComment(org, repo, number, comment string) bool {
	m.method = "CreateIssueComment"
	return m.successfulCreateIssueComment
}

func (m *mockClient) AddIssueLabels(org, repo, number string, labels []string) bool {
	m.method = "AddIssueLabels"
	return m.successfulAddIssueLabels
}

func (m *mockClient) RemoveIssueLabels(org, repo, number string, labels []string) bool {
	m.method = "RemoveIssueLabels"
	return m.successfulRemoveIssueLabels
}

func (m *mockClient) AddPRLabels(org, repo, number string, labels []string) bool {
	m.method = "AddPRLabels"
	return m.successfulAddPRLabels
}

func (m *mockClient) RemovePRLabels(org, repo, number string, labels []string) bool {
	m.method = "RemovePRLabels"
	return m.successfulRemovePRLabels
}

func (m *mockClient) CheckIfPRCreateEvent(evt *client.GenericEvent) bool {
	m.method = "CheckIfPRCreateEvent"
	return m.successfulCheckIfPRCreateEvent
}

func (m *mockClient) CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) bool {
	m.method = "CheckIfPRSourceCodeUpdateEvent"
	return m.successfulCheckIfPRSourceCodeUpdateEvent
}

func (m *mockClient) GetPullRequestCommits(org, repo, number string) ([]client.PRCommit, bool) {
	m.method = "GetPullRequestCommits"
	return m.commits, m.successfulGetPullRequestCommits
}

func (m *mockClient) GetPullRequestLabels(org, repo, number string) ([]string, bool) {
	m.method = "GetPullRequestLabels"
	return m.labels, m.successfulGetPullRequestLabels
}

func (m *mockClient) GetIssueLabels(org, issueID string) ([]string, bool) {
	m.method = "GetIssueLabels"
	return m.labels, m.successfulGetIssueLabels
}

func (m *mockClient) GetRepoIssueLabels(org, repo string) ([]string, bool) {
	m.method = "GetRepoIssueLabels"
	return m.labels, m.successfulGetRepoIssueLabels
}

const (
	org       = "org1"
	repo      = "repo1"
	number    = "1"
	commenter = "commenter1"
	label     = "label1"
)

func TestRemovePRLabels(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		CommentUpdateLabelFailed: "%s, 1123, %s",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	case1 := "No labels to remove"
	cli.method = case1
	// No labels to remove
	bot.removePRLabels(org, repo, number, commenter, []string{})
	assert.Equal(t, case1, cli.method)

	case2 := "RemovePRLabels"
	cli.method = case2
	cli.successfulRemovePRLabels = true
	// Successfully remove labels
	bot.removePRLabels(org, repo, number, commenter, []string{label})
	assert.Equal(t, case2, cli.method)

	case3 := "CreatePRComment"
	cli.method = case3
	cli.successfulRemovePRLabels = false
	// Failed to remove labels
	bot.removePRLabels(org, repo, number, commenter, []string{label})
	assert.Equal(t, case3, cli.method)

}

func TestAddLabels(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		CommentUpdateLabelFailed: "%s, 1123, %s",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	case1 := "No labels to add"
	cli.method = case1
	// No labels to add
	bot.addPRLabels(org, repo, number, commenter, []string{})
	assert.Equal(t, case1, cli.method)

	case2 := "AddPRLabels"
	cli.method = case2
	cli.successfulAddPRLabels = true
	// Successfully add labels
	bot.addPRLabels(org, repo, number, commenter, []string{label})
	assert.Equal(t, case2, cli.method)

	case3 := "CreatePRComment"
	cli.method = case3
	cli.successfulAddPRLabels = false
	// Failed to add labels
	bot.addPRLabels(org, repo, number, commenter, []string{label})
	assert.Equal(t, case3, cli.method)

}

func TestClearLabelWhenPRSourceCodeUpdated(t *testing.T) {

	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{
		CommentRemoveLabelsWhenPRSourceCodeUpdated: "1123, %s",
	}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	case1 := "CheckIfPRSourceCodeUpdateEvent"
	cli.method = case1
	cli.successfulCheckIfPRSourceCodeUpdateEvent = false
	cnf := &repoConfig{}
	// Not a pull request source code update event
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case1, cli.method)

	cli.successfulCheckIfPRSourceCodeUpdateEvent = true
	// No labels to clear
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case1, cli.method)

	cnf.ClearLabels = []string{label}
	case3 := "GetPullRequestLabels"
	cli.method = case3
	cli.successfulGetPullRequestLabels = false
	// there is no labels in the PR
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case3, cli.method)

	cli.successfulGetPullRequestLabels = true
	cli.labels = []string{label + "1"}
	// there is no intersection between cleared labels and PR's labels
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case3, cli.method)

	cli.labels = []string{label}
	case5 := "RemovePRLabels"
	cli.method = case5
	cli.successfulRemovePRLabels = false
	// there is a intersection between cleared labels and PR's labels
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case5, cli.method)

	cli.successfulRemovePRLabels = true
	case6 := "CreatePRComment"
	cli.method = case6
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, cnf, &client.GenericEvent{})
	assert.Equal(t, case6, cli.method)
}

func TestHandleSquashLabel(t *testing.T) {
	mc := new(mockClient)
	bot := &robot{cli: mc, cnf: &configuration{}}

	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)
	case1 := "CreatePRComment"
	cli.method = case1
	cli.successfulGetPullRequestCommits = false
	cnf := &repoConfig{}
	// The PR has no commits
	bot.handleSquashLabel(org, repo, number, cnf)
	assert.Equal(t, case1, cli.method)

	case2 := "GetPullRequestCommits"
	cli.method = case2
	cli.successfulGetPullRequestCommits = true
	cnf.UnableCheckingSquash = true
	// the PR squash check is disable
	bot.handleSquashLabel(org, repo, number, cnf)
	assert.Equal(t, case2, cli.method)

	cnf.UnableCheckingSquash = false
	case3 := "GetPullRequestLabels"
	cli.method = case3
	cli.commits = []client.PRCommit{
		{
			AuthorName: "user1",
		},
		{
			AuthorName: "user2",
		},
	}
	cnf.CommitsThreshold = 1
	bot.cnf.SquashCommitLabel = "squash"
	cli.successfulGetPullRequestLabels = true
	cli.labels = []string{bot.cnf.SquashCommitLabel}
	// the PR squash check is able, commits number is larger than threshold, but PR's labels already contains squash label
	bot.handleSquashLabel(org, repo, number, cnf)
	assert.Equal(t, case3, cli.method)

	case4 := "AddPRLabels"
	cli.method = case4
	// the PR squash check is able, commits number is larger than threshold, but PR's labels not contains squash label
	cli.labels = []string{"sig/aaa"}
	bot.handleSquashLabel(org, repo, number, cnf)
	assert.Equal(t, case4, cli.method)

	case5 := "RemovePRLabels"
	cli.method = case5
	cnf.CommitsThreshold = 2
	// the PR squash check is able, commits number is within than threshold, but PR's labels contains squash label
	cli.labels = []string{bot.cnf.SquashCommitLabel}
	bot.handleSquashLabel(org, repo, number, cnf)
	assert.Equal(t, case5, cli.method)
}
