// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package templates

import (
	"fmt"
	"html/template"
	"strings"

	"code.gitea.io/gitea/modules/base"
	"github.com/dustin/go-humanize"
)

type StringUtils struct{}

var stringUtils = StringUtils{}

func NewStringUtils() *StringUtils {
	return &stringUtils
}

func (su *StringUtils) ToString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case template.HTML:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
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

func (su *StringUtils) Cut(s, sep string) []any {
	before, after, found := strings.Cut(s, sep)
	return []any{before, after, found}
}

func (su *StringUtils) EllipsisString(s string, max int) string {
	return base.EllipsisString(s, max)
}

func (su *StringUtils) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// FormatFileSize use in dvc file size
func (su *StringUtils) FormatFileSize(fileSize *uint64) string {
	return humanize.Bytes(*fileSize)
}
