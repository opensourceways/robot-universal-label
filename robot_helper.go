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
	"k8s.io/apimachinery/pkg/util/sets"
	"net/url"
	"slices"
	"strings"
)

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

func (bot *robot) clearLabelWhenPRSourceCodeUpdated(org, repo, number string, repoCnf *repoConfig, evt *client.GenericEvent) {
	if !bot.cli.CheckIfPRSourceCodeUpdateEvent(evt) {
		return
	}

	clearLabelSet := sets.New[string](repoCnf.ClearLabels...)
	if clearLabelSet.Len() == 0 && repoCnf.clearLabelsByRegexp == nil {
		return
	}

	prLabels, _ := bot.cli.GetPullRequestLabels(org, repo, number)
	if len(prLabels) == 0 {
		return
	}

	for _, l := range prLabels {
		if repoCnf.clearLabelsByRegexp != nil && repoCnf.clearLabelsByRegexp.MatchString(l) {
			clearLabelSet.Insert(l)
		}
	}

	prLabelSet := sets.New[string](prLabels...)
	clearLabels := prLabelSet.Intersection(clearLabelSet).UnsortedList()
	if len(clearLabels) == 0 {
		return
	}

	if bot.cli.RemovePRLabels(org, repo, number, clearLabels) {
		comment := fmt.Sprintf(bot.cnf.CommentRemoveLabelsWhenPRSourceCodeUpdated, strings.Join(clearLabels, ", "))
		bot.cli.CreatePRComment(org, repo, number, comment)
	}
}

func (bot *robot) addLabels(org, repo, number, commenter string, addLabels []string) {
	if len(addLabels) == 0 {
		return
	}

	success := bot.cli.AddPRLabels(org, repo, number, addLabels)
	if !success {
		comment := fmt.Sprintf(bot.cnf.CommentUpdateLabelFailed, commenter, strings.Join(addLabels, ", "))
		bot.cli.CreatePRComment(org, repo, number, comment)
	}
}

func (bot *robot) removeLabels(org, repo, number, commenter string, removeLabels []string) {
	if len(removeLabels) == 0 {
		return
	}

	success := bot.cli.RemovePRLabels(org, repo, number, removeLabels)
	if !success {
		comment := fmt.Sprintf(bot.cnf.CommentUpdateLabelFailed, commenter, strings.Join(removeLabels, ", "))
		bot.cli.CreatePRComment(org, repo, number, comment)
	}
}
