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

import "regexp"

var (
	regexpCommentByAnyoneToAddLabel  = regexp.MustCompile(`(?m)^/(kind|priority|sig|good)\s*(.*?)\s*$`)
	regexpCommentByAnyoneRemoveLabel = regexp.MustCompile(`(?m)^/remove-(kind|priority|sig|good)\s*(.*?)\s*$`)
)

func getMatchedLabels(comment *string) ([]string, []string) {
	return parseLabels(comment, commonLabelRegex),
		parseLabels(comment, removeCommonLabelRegex)
}

func matchLabelFromCommentLine(comment string, reg *regexp.Regexp) []string {
	var labels []string
	r := reg.FindAllStringSubmatch(comment, -1)
	for _, v := range r {
		if len(v) < 3 {
			continue
		}

		for _, item := range v[2:] {
			if v[1] == "good" {
				prefix := v[1]
				labels = append(labels, prefix+item)
				continue
			}
			prefix := v[1] + "/"
			labels = append(labels, prefix+item)
		}
	}

	return labels
}
