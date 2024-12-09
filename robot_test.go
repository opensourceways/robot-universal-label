// Copyright 2024 Chao Feng
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
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func botHelper(t *testing.T) (*robot, *mockClient) {
	mc := new(mockClient)

	logger := framework.NewLogger().WithField("component", component)
	cnf := &configuration{}
	err := utils.LoadFromYaml(findTestdata(t, "config.yaml"), cnf)
	assert.Equal(t, nil, err)
	bot := &robot{cli: mc, cnf: cnf, log: logger}
	cli, ok := bot.cli.(*mockClient)
	assert.Equal(t, true, ok)

	assert.Equal(t, cnf, bot.GetConfigmap())
	assert.Equal(t, logger, bot.GetLogger())

	return bot, cli
}

func TestHandlePullRequestEvent(t *testing.T) {
	bot, cli := botHelper(t)

	evtOrg, evtRepo, evtNo := org, repo, number
	evt := &client.GenericEvent{
		Org:    &evtOrg,
		Repo:   &evtRepo,
		Number: &evtNo,
	}

	case1 := "No org or repo matched in the config"
	cli.method = case1
	// No org or repo matched in the config
	bot.handlePullRequestEvent(evt, nil, bot.log)
	assert.Equal(t, case1, cli.method)

	evtOrg = "owner1"
	cli.successfulCheckIfPRCreateEvent = false
	cli.successfulCheckIfPRSourceCodeUpdateEvent = false
	case2 := "CheckIfPRSourceCodeUpdateEvent"
	cli.method = case2
	// Org matched, but event is not within the scope of processing
	bot.handlePullRequestEvent(evt, nil, bot.log)
	assert.Equal(t, case2, cli.method)

	cli.successfulCheckIfPRCreateEvent = true
	cli.successfulGetPullRequestCommits = false
	cli.successfulCheckIfPRSourceCodeUpdateEvent = false
	case3 := "CheckIfPRSourceCodeUpdateEvent"
	cli.method = case3
	// Org matched, and event is handle over
	bot.handlePullRequestEvent(evt, nil, bot.log)
	assert.Equal(t, case3, cli.method)
}

func TestHandleCommentEvent(t *testing.T) {
	bot, cli := botHelper(t)

	evtOrg, evtRepo, evtNo, evtCommenter := org, repo, number, "user3"
	evt := &client.GenericEvent{
		Org:       &evtOrg,
		Repo:      &evtRepo,
		Number:    &evtNo,
		Commenter: &evtCommenter,
	}

	case1 := "No org or repo matched in the config"
	cli.method = case1
	// No org or repo matched in the config
	bot.handlePullRequestCommentEvent(evt, nil, bot.log)
	assert.Equal(t, case1, cli.method)

	evtOrg, evtRepo = "owner2", "repo1"
	evtComment := "/kind bug \n /remove-kind bug"
	evt.Comment = &evtComment
	case2 := "CreatePRComment"
	cli.method = case2
	// the same label is to add and to delete in the command line
	bot.handlePullRequestCommentEvent(evt, nil, bot.log)
	assert.Equal(t, case2, cli.method)

	evtComment = "/kind bug"
	evt.Comment = &evtComment
	case3 := "GetPullRequestLabels"
	cli.labels = []string{"kind/bug"}
	cli.method = case3
	bot.handlePullRequestCommentEvent(evt, nil, bot.log)
	assert.Equal(t, case3, cli.method)
}
