package main

import (
	"context"
	"log"
)

type Runner struct {
	Config  Config
	Fetcher Fetcher
}

func NewRunner(c Config) *Runner {
	ctx := context.Background()
	f := &fetcher{
		Repo:  c.Repo,
		Token: c.Token,
	}
	f.Auth(ctx)
	r := &Runner{
		Config:  c,
		Fetcher: f,
	}

	return r
}

func (r *Runner) Run(isDryRun bool) error {
	rel := &Release{
		Fetcher: r.Fetcher,
		Config:  r.Config,
	}
	ctx := context.Background()
	bpr, err := rel.Build(ctx)
	if err != nil {
		return err
	}

	number, isOpened, err := r.Fetcher.OpenedReleasePRNumber(ctx)
	if err != nil {
		return err
	}
	if isOpened {
		log.Println("[DEBUG] update pull request")
	} else {
		log.Println("[DEBUG] create pull request")
	}
	log.Printf("[DEBUG] title=%s", bpr.Title)
	log.Printf("[DEBUG] %s", bpr.Body)

	if isDryRun {
		return nil
	}

	if isOpened {
		_, err := r.Fetcher.UpdatePullRequest(ctx, bpr.Title, bpr.Body, number, r.Config.Labels)
		if err != nil {
			return err
		}
	} else {
		_, err := r.Fetcher.CreatePullRequest(ctx, bpr.Title, bpr.Body, r.Config.Labels)
		if err != nil {
			return err
		}
	}

	return nil
}
