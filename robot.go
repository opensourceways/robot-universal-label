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
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/robot-framework-lib/utils"
	"github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"slices"
)

// iClient is an interface that defines methods for client-side interactions
type iClient interface {
	// CreatePRComment creates a comment for a pull request in a specified organization and repository
	CreatePRComment(org, repo, number, comment string) (success bool)
	// CreateIssueComment creates a comment for an issue in a specified organization and repository
	CreateIssueComment(org, repo, number, comment string) (success bool)
	// CheckPermission checks the permission of a user for a specified repository
	CheckPermission(org, repo, username string) (pass, success bool)
	// UpdateIssue updates the state of an issue in a specified organization and repository
	UpdateIssue(org, repo, number, state string) (success bool)
	// UpdatePR updates the state of a pull request in a specified organization and repository
	UpdatePR(org, repo, number, state string) (success bool)
	// GetIssueLinkedPRNumber retrieves the number of a pull request linked to a specified issue
	GetIssueLinkedPRNumber(org, repo, number string) (num int, success bool)

	CreateRepoIssueLabel(org, repo, name, color string) (success bool)
	DeleteRepoIssueLabel(org, repo, name string) (success bool)
	AddIssueLabels(org, repo, number string, labels []string) (success bool)
	RemoveIssueLabels(org, repo, number string, labels []string) (success bool)
	AddPRLabels(org, repo, number string, labels []string) (success bool)
	RemovePRLabels(org, repo, number string, labels []string) (success bool)
	CheckIfPRCreateEvent(evt *client.GenericEvent) (yes bool)
	CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) (yes bool)
	GetPullRequestCommits(org, repo, number string) (result []client.PRCommit, success bool)
	GetPullRequestLabels(org, repo, number string) (result []string, success bool)
}

type robot struct {
	cli iClient
	cnf *configuration
	log *logrus.Entry
}

func newRobot(c *configuration, token []byte) *robot {
	logger := framework.NewLogger().WithField("component", component)
	return &robot{cli: client.NewClient(token, logger), cnf: c, log: logger}
}

func (bot *robot) GetConfigmap() config.Configmap {
	return bot.cnf
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterPullRequestHandler(bot.handlePullRequestEvent)
	p.RegisterIssueCommentHandler(bot.handleCommentEvent)
	p.RegisterPullRequestCommentHandler(bot.handleCommentEvent)
}

func (bot *robot) GetLogger() *logrus.Entry {
	return bot.log
}

var (
	// the value from configuration.EventStateOpened
	eventStateOpened = "opened"
	// the value from configuration.EventStateClosed
	eventStateClosed = "closed"
	// the value from configuration.CommentNoPermissionOperateIssue
	commentNoPermissionOperateIssue = ""
	// the value from configuration.CommentIssueNeedsLinkPR
	commentIssueNeedsLinkPR = ""
	// the value from configuration.CommentListLinkingPullRequestsFailure
	commentListLinkingPullRequestsFailure = ""
	// the value from configuration.CommentNoPermissionOperatePR
	commentNoPermissionOperatePR = ""
)

const (
	// placeholderCommenter is a placeholder string for the commenter's name
	placeholderCommenter = "__commenter__"
	// placeholderAction is a placeholder string for the action
	placeholderAction = "__action__"
)

var (
	// regexpReopenComment is a compiled regular expression for reopening comments
	regexpReopenComment = regexp.MustCompile(`(?mi)^/reopen\s*$`)
	// regexpCloseComment is a compiled regular expression for closing comments
	regexpCloseComment = regexp.MustCompile(`(?mi)^/close\s*$`)
)

func (bot *robot) handlePullRequestEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf := bot.cnf.get(org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the warning and returns
	if repoCnf == nil {
		logger.Warningf("no config for the repo: " + org + "/" + repo)
		return
	}

	// Checks if PR is firstly created or PR source code is updated
	if !(bot.cli.CheckIfPRCreateEvent(evt) || bot.cli.CheckIfPRSourceCodeUpdateEvent(evt)) {
		return
	}

	bot.handleSquashLabel(org, repo, number, repoCnf)
}

func (bot *robot) handleSquashLabel(org, repo, number string, repoCnf *repoConfig) {
	commits, success := bot.cli.GetPullRequestCommits(org, repo, number)
	if !success {
		bot.cli.CreatePRComment(org, repo, number, bot.cnf.CommentCommandTrigger)
		return
	}

	if repoCnf.UnableCheckingSquash == false {
		prLabels, _ := bot.cli.GetPullRequestLabels(org, repo, number)
		if uint(len(commits)) > repoCnf.CommitsThreshold && !slices.Contains(prLabels, bot.cnf.SquashCommitLabel) {
			bot.cli.AddPRLabels(org, repo, number, []string{bot.cnf.SquashCommitLabel})
		}

		if uint(len(commits)) <= repoCnf.CommitsThreshold && slices.Contains(prLabels, bot.cnf.SquashCommitLabel) {
			bot.cli.RemovePRLabels(org, repo, number, []string{url.QueryEscape(bot.cnf.SquashCommitLabel)})
		}
	}
}

func (bot *robot) handleCommentEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf := bot.cnf.get(org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the warning and returns
	if repoCnf == nil {
		logger.Warningf("no config for the repo: " + org + "/" + repo)
		return
	}

}
