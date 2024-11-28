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
	"errors"
	"github.com/opensourceways/server-common-lib/config"
	"regexp"
)

// configuration holds a list of repoConfig configurations.
type configuration struct {
	ConfigItems []repoConfig `json:"config_items,omitempty"`
}

// Validate to check the configmap data's validation, returns an error if invalid
func (c *configuration) Validate() error {
	if c == nil {
		return errors.New("configuration is nil")
	}

	// Validate each repo configuration
	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

// get retrieves a repoConfig for a given organization and repository.
// Returns the repoConfig if found, otherwise returns nil.
func (c *configuration) get(org, repo string) *repoConfig {
	if c == nil || len(c.ConfigItems) == 0 {
		return nil
	}

	for i := range c.ConfigItems {
		ok, _ := c.ConfigItems[i].RepoFilter.CanApply(org, org+"/"+repo)
		if ok {
			return &c.ConfigItems[i]
		}
	}

	return nil
}

// repoConfig is a configuration struct for a organization and repository.
// It includes a RepoFilter and a boolean value indicating if an issue can be closed only when its linking PR exists.
type repoConfig struct {
	// RepoFilter is used to filter repositories.
	config.RepoFilter
	// ClearLabels specifies labels that should be removed when the codes of PR are changed.
	ClearLabels []string `json:"clear_labels,omitempty"`

	// ClearLabelsByRegexp specifies a expression which can match a list of labels that
	// should be removed when the codes of PR are changed.
	ClearLabelsByRegexp string `json:"clear_labels_by_regexp,omitempty"`
	clearLabelsByRegexp *regexp.Regexp

	// AllowCreatingLabelsByCollaborator is a tag which will lead to create unavailable labels
	// by collaborator if it is true.
	AllowCreatingLabelsByCollaborator bool `json:"allow_creating_labels_by_collaborator,omitempty"`

	SquashConfig
}

// validate to check the repoConfig data's validation, returns an error if invalid
func (c *repoConfig) validate() error {
	// If the bot is not configured to monitor any repositories, return an error.
	if len(c.Repos) == 0 {
		return errors.New("the repositories configuration can not be empty")
	}

	return c.RepoFilter.Validate()
}

type SquashConfig struct {
	// UnableCheckingSquash indicates whether unable checking squash.
	UnableCheckingSquash bool `json:"unable_checking_squash,omitempty"`

	// CommitsThreshold Check the threshold of the number of PR commits,
	// and add the label specified by SquashCommitLabel to the PR if this value is exceeded.
	// zero means no check.
	CommitsThreshold uint `json:"commits_threshold,omitempty"`

	// SquashCommitLabel Specify the label whose PR exceeds the threshold. default: stat/needs-squash
	SquashCommitLabel string `json:"squash_commit_label,omitempty"`
}
