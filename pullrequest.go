package main

import (
	"strings"
)

type PullRequest struct {
	Number int
	Title  string
	Head   string
	Base   string
	Author string
}

type PullRequests []PullRequest

func (prs PullRequests) Titles() string {
	titles := make([]string, 0, len(prs))
	for _, pr := range prs {
		titles = append(titles, pr.Title)
	}

	return strings.Join(titles, " ")

}
