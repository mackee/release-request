package main

import (
	"strings"
	"time"

	"github.com/google/go-github/github"
)

type Branch struct {
	Name         string
	Commit       string
	Commiter     string
	CommiterDate time.Time
}

type Branches []Branch

func ToBranch(b *github.Branch) Branch {
	return Branch{
		Name:         *b.Name,
		Commit:       b.Commit.GetSHA(),
		Commiter:     b.Commit.Author.GetName(),
		CommiterDate: b.Commit.Commit.Committer.GetDate(),
	}
}

func (bs Branches) Name() string {
	ns := make([]string, 0, len(bs))
	for _, b := range bs {
		ns = append(ns, b.Name)
	}
	return strings.Join(ns, " ")
}
