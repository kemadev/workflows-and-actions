package main

import (
	"github.com/go-git/go-git/v5"
)

var GitRepoBasePath = GetGitBasePath()

func GetGitBasePath() string {
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return ""
	}
	o, err := r.Remote("origin")
	if err != nil {
		return ""
	}
	if len(o.Config().URLs) == 0 {
		return ""
	}
	return o.Config().URLs[0]
}
