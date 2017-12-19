package main

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type Fetcher interface {
	Auth(context.Context)
	OpenedReleasePRNumber(context.Context) (int, bool, error)
	MergedPullRequests(context.Context) (PullRequests, error)
	OpenedPullRequests(context.Context, string) (PullRequests, error)
	GetComparison() Comparison
	CreatePullRequest(context.Context, string, string, []string) (PullRequest, error)
	UpdatePullRequest(context.Context, string, string, int, []string) (PullRequest, error)
}

type fetcher struct {
	Repo       Repo
	Token      string
	client     *github.Client
	comparison *Comparison
}

func (f *fetcher) Auth(ctx context.Context) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: f.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	f.client = github.NewClient(tc)
}

func (f *fetcher) OpenedReleasePRNumber(ctx context.Context) (int, bool, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		Head:  f.Repo.JoinedHead(),
		Base:  f.Repo.Base,
	}
	prs, _, err := f.client.PullRequests.List(ctx, f.Repo.Owner, f.Repo.Repo, opts)
	if err != nil {
		return 0, false, errors.Wrap(err, "fail fetch pull requests")
	}
	if len(prs) == 0 {
		return 0, false, nil
	}

	return prs[0].GetNumber(), true, nil
}

func (f *fetcher) headBranch(ctx context.Context) (Branch, error) {
	b, _, err := f.client.Repositories.GetBranch(ctx, f.Repo.Owner, f.Repo.Repo, f.Repo.Head)
	if err != nil {
		return Branch{}, errors.Wrap(err, "fail fetching head branch information")
	}

	branch := ToBranch(b)
	return branch, nil
}

func (f *fetcher) baseBranch(ctx context.Context) (Branch, error) {
	b, _, err := f.client.Repositories.GetBranch(ctx, f.Repo.Owner, f.Repo.Repo, f.Repo.Base)
	if err != nil {
		return Branch{}, errors.Wrap(err, "fail fetching information of base branch")
	}

	branch := ToBranch(b)
	return branch, nil
}

func (f *fetcher) compare(ctx context.Context) (Comparison, error) {
	head, err := f.headBranch(ctx)
	if err != nil {
		return Comparison{}, err
	}
	base, err := f.baseBranch(ctx)
	if err != nil {
		return Comparison{}, err
	}

	comparison, _, err := f.client.Repositories.CompareCommits(
		ctx, f.Repo.Owner, f.Repo.Repo, base.Commit, head.Commit,
	)
	if err != nil {
		return Comparison{}, errors.Wrap(err, "fail fetching comapre commits")
	}

	var com Comparison
	if len(comparison.Files) > 0 {
		files := make([]ComparisonFile, 0, len(comparison.Files))
		for _, f := range comparison.Files {
			cf := ComparisonFile{
				Name:      f.GetFilename(),
				SHA:       f.GetSHA(),
				Deletions: f.GetDeletions(),
				Additions: f.GetAdditions(),
				Patch:     f.GetPatch(),
			}
			files = append(files, cf)
		}
		com.Files = files
	}
	if len(comparison.Commits) > 0 {
		commits := make([]ComparisonCommit, 0, len(comparison.Commits))
		for _, c := range comparison.Commits {
			var parentHashes []string
			if len(c.Parents) > 0 {
				parentHashes = make([]string, 0, len(c.Parents))
				for _, parent := range c.Parents {
					parentHashes = append(parentHashes, parent.GetSHA())
				}
			}
			cc := ComparisonCommit{
				Parents: parentHashes,
				Message: c.Commit.GetMessage(),
			}
			commits = append(commits, cc)
		}
		com.Commits = commits
	}

	return com, nil
}

func (f *fetcher) fetchCompare(ctx context.Context) error {
	if f.comparison != nil {
		return nil
	}

	comparison, err := f.compare(ctx)
	if err != nil {
		return err
	}

	f.comparison = &comparison
	return nil
}

func (f *fetcher) MergedPullRequests(ctx context.Context) (PullRequests, error) {
	err := f.fetchCompare(ctx)
	if err != nil {
		return nil, err
	}

	commits := f.comparison.Commits
	if commits == nil {
		return nil, nil
	}

	prs := make([]*PullRequest, 0, len(commits)/2)
	for _, commit := range commits {
		if !commit.IsMerge() || commit.IsPullFromMaster() {
			continue
		}
		pr, err := commit.PullRequest()
		if err != nil {
			log.Printf("[DEBUG] %s", err)
			continue
		}
		prs = append(prs, &pr)
	}

	eg := errgroup.Group{}
	for _, pr := range prs {
		_pr := pr
		eg.Go(func() error {
			err := f.FetchPullRequest(ctx, _pr)
			if err == nil {
				return nil
			}
			cause := errors.Cause(err)
			errResp, ok := cause.(*github.ErrorResponse)
			if ok && errResp.Response.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	freezedPRs := make(PullRequests, 0, len(prs))
	for _, pr := range prs {
		if !pr.Fetched {
			continue
		}
		if pr.Base != f.Repo.Head {
			continue
		}
		freezedPRs = append(freezedPRs, *pr)
	}

	return freezedPRs, nil
}

func (f *fetcher) FetchPullRequest(ctx context.Context, pr *PullRequest) error {
	prr, _, err := f.client.PullRequests.Get(ctx, f.Repo.Owner, f.Repo.Repo, pr.Number)
	if err != nil {
		return errors.Wrap(err, "fail fetching pull request")
	}

	pr.Head = prr.GetHead().GetRef()
	pr.Base = prr.GetBase().GetRef()
	pr.Title = prr.GetTitle()
	pr.Author = prr.GetUser().GetLogin()
	pr.Fetched = true

	return nil
}

func (f *fetcher) OpenedPullRequests(ctx context.Context, label string) (PullRequests, error) {
	opt := &github.IssueListByRepoOptions{
		State:  "open",
		Labels: []string{label},
	}
	issues, _, err := f.client.Issues.ListByRepo(ctx, f.Repo.Owner, f.Repo.Repo, opt)
	if err != nil {
		return nil, errors.Wrap(err, "fail fetching issues on OpenedPullRequests")
	}

	prs := make([]*PullRequest, 0, len(issues))
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}

		pr := &PullRequest{
			Number: issue.GetNumber(),
			Title:  issue.GetTitle(),
		}
		prs = append(prs, pr)
	}

	eg := errgroup.Group{}
	for _, pr := range prs {
		_pr := pr
		eg.Go(func() error {
			return f.FetchPullRequest(ctx, _pr)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	freezedPRs := make(PullRequests, 0, len(prs))
	for _, pr := range prs {
		if pr.Base != f.Repo.Head {
			continue
		}
		freezedPRs = append(freezedPRs, *pr)
	}

	return freezedPRs, nil
}

func (f *fetcher) GetComparison() Comparison {
	return *f.comparison
}

func (f *fetcher) CreatePullRequest(ctx context.Context, title, body string, labels []string) (PullRequest, error) {
	pull := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &f.Repo.Head,
		Base:  &f.Repo.Base,
	}
	pr, _, err := f.client.PullRequests.Create(ctx, f.Repo.Owner, f.Repo.Repo, pull)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "fail create pull request")
	}

	ppr := PullRequest{
		Number: pr.GetNumber(),
		Title:  pr.GetTitle(),
		Head:   pr.GetHead().GetRef(),
		Base:   pr.GetBase().GetRef(),
		Author: pr.GetUser().GetLogin(),
	}

	err = f.updateLabel(ctx, ppr.Number, labels)
	if err != nil {
		return PullRequest{}, err
	}

	return ppr, nil
}

func (f *fetcher) updateLabel(ctx context.Context, number int, labels []string) error {
	if len(labels) == 0 {
		return nil
	}
	_, _, err := f.client.Issues.AddLabelsToIssue(ctx, f.Repo.Owner, f.Repo.Repo, number, labels)
	if err != nil {
		return errors.Wrap(err, "fail update label to issue")
	}

	return nil
}

func (f *fetcher) UpdatePullRequest(ctx context.Context, title, body string, number int, labels []string) (PullRequest, error) {
	pull := &github.PullRequest{
		Title: &title,
		Body:  &body,
	}
	pr, _, err := f.client.PullRequests.Edit(ctx, f.Repo.Owner, f.Repo.Repo, number, pull)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "fail update pull request")
	}

	ppr := PullRequest{
		Number: pr.GetNumber(),
		Title:  pr.GetTitle(),
		Head:   pr.GetHead().GetRef(),
		Base:   pr.GetBase().GetRef(),
		Author: pr.GetUser().GetLogin(),
	}

	err = f.updateLabel(ctx, ppr.Number, labels)
	if err != nil {
		return PullRequest{}, err
	}

	return ppr, nil
}
