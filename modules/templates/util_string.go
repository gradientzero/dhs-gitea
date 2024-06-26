// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package templates

import (
	"github.com/dustin/go-humanize"
	"strings"

	"code.gitea.io/gitea/modules/base"
)

type StringUtils struct{}

var stringUtils = StringUtils{}

func NewStringUtils() *StringUtils {
	return &stringUtils
}

func (su *StringUtils) HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (su *StringUtils) Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (su *StringUtils) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (su *StringUtils) Join(a []string, sep string) string {
	return strings.Join(a, sep)
}

func (su *StringUtils) Concat(str ...string) string {
	output := ""
	for _, v := range str {
		output += v
	}
	return output
}

func (su *StringUtils) Cut(s, sep string) []any {
	before, after, found := strings.Cut(s, sep)
	return []any{before, after, found}
}

func (su *StringUtils) EllipsisString(s string, max int) string {
	return base.EllipsisString(s, max)
}

// FormatFileSize use in dvc file size
func (su *StringUtils) FormatFileSize(fileSize *uint64) string {
	return humanize.Bytes(*fileSize)
}
