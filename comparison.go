package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Comparison struct {
	Files   []ComparisonFile   `json:"files"`
	Commits []ComparisonCommit `json:"commits"`
}

type ComparisonFile struct {
	Name      string `json:"name"`
	SHA       string `json:"sha"`
	Deletions int    `json:"deletions"`
	Additions int    `json:"additions"`
	Patch     string `json:"patch"`
}

type ComparisonCommit struct {
	Parents []string `json:"parents"`
	Message string   `json:"message"`
}

func (c ComparisonCommit) IsMerge() bool {
	return len(c.Parents) > 1
}

func (c ComparisonCommit) IsPullFromMaster() bool {
	return strings.HasPrefix(c.Message, "Merge branch 'master' into")
}

var prMergeReg = regexp.MustCompile(`^(?:[a-f0-9]+ )?Merge pull request #([0-9]+) from \S+`)

func (c ComparisonCommit) PullRequest() (PullRequest, error) {
	var pr PullRequest

	if !c.IsMerge() {
		return pr, errors.New("not a merge commit")
	}

	matches := prMergeReg.FindStringSubmatch(c.Message)
	if len(matches) <= 1 {
		return pr, fmt.Errorf("not a stricted commit message: %s", c.Message)
	}
	numberStr := matches[1]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return pr, errors.Wrap(err, "cannot convert to integer at Pull Request number")
	}

	pr = PullRequest{
		Number: number,
	}

	return pr, nil
}
