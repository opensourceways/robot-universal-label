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
	"fmt"
	"github.com/opensourceways/robot-framework-lib/client"
	"github.com/opensourceways/robot-framework-lib/config"
	"github.com/opensourceways/robot-framework-lib/framework"
	"github.com/opensourceways/robot-framework-lib/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

// iClient is an interface that defines methods for client-side interactions
type iClient interface {
	// CreatePRComment creates a comment for a pull request in a specified organization and repository
	CreatePRComment(org, repo, number, comment string) (success bool)
	CreateIssueComment(org, repo, number, comment string) (success bool)
	AddIssueLabels(org, repo, number string, labels []string) (success bool)
	RemoveIssueLabels(org, repo, number string, labels []string) (success bool)
	AddPRLabels(org, repo, number string, labels []string) (success bool)
	RemovePRLabels(org, repo, number string, labels []string) (success bool)
	CheckIfPRCreateEvent(evt *client.GenericEvent) (yes bool)
	CheckIfPRSourceCodeUpdateEvent(evt *client.GenericEvent) (yes bool)
	GetPullRequestCommits(org, repo, number string) (result []client.PRCommit, success bool)
	GetPullRequestLabels(org, repo, number string) (result []string, success bool)
	GetIssueLabels(org, issueID string) (result []string, success bool)
	GetRepoIssueLabels(org, repo string) (result []string, success bool)
	CheckPermission(org, repo, username string) (pass, success bool)
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
	p.RegisterIssueCommentHandler(bot.handleIssueCommentEvent)
	p.RegisterPullRequestCommentHandler(bot.handlePullRequestCommentEvent)
}

func (bot *robot) GetLogger() *logrus.Entry {
	return bot.log
}

const (
	logWarningMessage = "no config for the repo: "
)

func (bot *robot) handlePullRequestEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf := bot.cnf.getRepoConfig(org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the warning and returns
	if repoCnf == nil {
		logger.Warning(logWarningMessage + org + "/" + repo)
		return
	}

	// Checks if PR is firstly created or PR source code is updated
	if !(bot.cli.CheckIfPRCreateEvent(evt) || bot.cli.CheckIfPRSourceCodeUpdateEvent(evt)) {
		return
	}

	bot.handleSquashLabel(org, repo, number, repoCnf)
	bot.clearLabelWhenPRSourceCodeUpdated(org, repo, number, repoCnf, evt)
}

func (bot *robot) handleIssueCommentEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf := bot.cnf.getRepoConfig(org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the warning and returns
	if repoCnf == nil {
		logger.Warning(logWarningMessage + org + "/" + repo)
		return
	}

	commenter := utils.GetString(evt.Commenter)
	commenter = strings.ReplaceAll(bot.cnf.UserMarkFormat, bot.cnf.PlaceholderCommenter, commenter)
	addLabels, removeLabels := matchLabels(utils.GetString(evt.Comment))
	if conflict, conflictLabels := checkIntersection(addLabels, removeLabels); conflict {
		comment := fmt.Sprintf(bot.cnf.CommentLabelCommandConflict, commenter, conflictLabels)
		bot.cli.CreateIssueComment(org, repo, number, comment)
		return
	}

	repoLabels, _ := bot.cli.GetRepoIssueLabels(org, repo)
	repoLabelSet := sets.New[string](repoLabels...)
	addLabelSet := sets.New[string](addLabels...)
	missingLabels := addLabelSet.Difference(repoLabelSet).UnsortedList()
	pass, _ := bot.cli.CheckPermission(org, repo, utils.GetString(evt.Commenter))
	if !pass && len(missingLabels) != 0 {
		bot.cli.CreateIssueComment(org, repo, number,
			fmt.Sprintf(bot.cnf.CommentAddNotExistLabel, commenter, strings.Join(missingLabels, ", ")))
		return
	}

	removeLabelSet := sets.New[string](removeLabels...)
	issueLabels, _ := bot.cli.GetIssueLabels(org, utils.GetString(evt.ID))
	issueLabelSet := sets.New[string](issueLabels...)
	bot.addIssueLabels(org, repo, number, commenter, addLabelSet.Difference(issueLabelSet).UnsortedList())
	bot.removeIssueLabels(org, repo, number, commenter, issueLabelSet.Intersection(removeLabelSet).UnsortedList())
}

func (bot *robot) handlePullRequestCommentEvent(evt *client.GenericEvent, cnf config.Configmap, logger *logrus.Entry) {
	org, repo, number := utils.GetString(evt.Org), utils.GetString(evt.Repo), utils.GetString(evt.Number)
	repoCnf := bot.cnf.getRepoConfig(org, repo)
	// If the specified repository not match any repository  in the repoConfig list, it logs the warning and returns
	if repoCnf == nil {
		logger.Warning(logWarningMessage + org + "/" + repo)
		return
	}

	commenter := utils.GetString(evt.Commenter)
	commenter = strings.ReplaceAll(bot.cnf.UserMarkFormat, bot.cnf.PlaceholderCommenter, commenter)
	addLabels, removeLabels := matchLabels(utils.GetString(evt.Comment))
	if conflict, conflictLabels := checkIntersection(addLabels, removeLabels); conflict {
		comment := fmt.Sprintf(bot.cnf.CommentLabelCommandConflict, commenter, conflictLabels)
		bot.cli.CreatePRComment(org, repo, number, comment)
		return
	}

	repoLabels, _ := bot.cli.GetRepoIssueLabels(org, repo)
	repoLabelSet := sets.New[string](repoLabels...)
	addLabelSet := sets.New[string](addLabels...)
	missingLabels := addLabelSet.Difference(repoLabelSet).UnsortedList()
	pass, _ := bot.cli.CheckPermission(org, repo, utils.GetString(evt.Commenter))
	if !pass && len(missingLabels) != 0 {
		bot.cli.CreatePRComment(org, repo, number,
			fmt.Sprintf(bot.cnf.CommentAddNotExistLabel, commenter, strings.Join(missingLabels, ", ")))
		return
	}

	removeLabelSet := sets.New[string](removeLabels...)
	prLabels, _ := bot.cli.GetPullRequestLabels(org, repo, number)
	prLabelSet := sets.New[string](prLabels...)
	bot.addPRLabels(org, repo, number, commenter, addLabelSet.Difference(prLabelSet).UnsortedList())
	bot.removePRLabels(org, repo, number, commenter, prLabelSet.Intersection(removeLabelSet).UnsortedList())
}
