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
	"k8s.io/apimachinery/pkg/util/sets"
	"regexp"
	"slices"
	"strings"
)

var (
	regexpCommentByAnyoneToAddLabel  = regexp.MustCompile(`^/(kind|priority|sig|good)[\t ]+[A-Za-z0-9_-]+$`)
	regexpCommentByAnyoneRemoveLabel = regexp.MustCompile(`^/remove-(kind|priority|sig|good)[\t ]+[A-Za-z0-9_-]+$`)
)

func matchLabels(comment string) (add []string, remove []string) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		label := matchLabelFromCommentLine(line, regexpCommentByAnyoneToAddLabel)
		if label != "" {
			add = append(add, label)
		}
		label = matchLabelFromCommentLine(line, regexpCommentByAnyoneRemoveLabel)
		if label != "" {
			remove = append(remove, label)
		}
	}

	return
}

func matchLabelFromCommentLine(oneLineComment string, reg *regexp.Regexp) string {
	var label string
	oneLineComment = strings.TrimSpace(oneLineComment)
	oneLineCommentByte := []byte(oneLineComment)
	index := reg.FindSubmatchIndex(oneLineCommentByte)
	if len(index) != 0 {
		splitIndex := index[len(index)-1]
		oneLineCommentByte[0] = oneLineCommentByte[0] ^ oneLineCommentByte[splitIndex]
		oneLineCommentByte[splitIndex] = oneLineCommentByte[0] ^ oneLineCommentByte[splitIndex]
		oneLineCommentByte[0] = oneLineCommentByte[0] ^ oneLineCommentByte[splitIndex]
		label = strings.ReplaceAll(strings.TrimSpace(string(oneLineCommentByte)), " ", "")
		label = strings.ReplaceAll(label, "\t", "")
	}

	if strings.HasPrefix(label, "remove-") {
		label = label[7:]
	}

	return label
}

func checkIntersection(add, remove []string) (bool, string) {
	if len(add) == 0 || len(remove) == 0 {
		return false, ""
	}

	addSet, removeSet := sets.Set[string]{}, sets.Set[string]{}
	for _, s := range add {
		addSet.Insert(strings.ToLower(s))
	}
	for _, s := range remove {
		removeSet.Insert(strings.ToLower(s))
	}

	list := addSet.Intersection(removeSet).UnsortedList()
	slices.Sort(list)
	if len(list) == 0 {
		return false, ""
	}

	return true, strings.Join(list, "**, **")
}
