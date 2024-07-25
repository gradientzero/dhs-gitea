// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import "time"

type Remote struct {
	Name       string
	Url        string
	AuthorName string
	DateAdded  time.Time
	Link       string
}

type RemoteGitBlame struct {
	AuthorName string
	Date       time.Time
}

type File struct {
	// https://stackoverflow.com/questions/10998222/json-parsing-of-int64-in-go-null-values
	// use *int64 instead of int64 to check null value
	Size *uint64
	Path string
}
