package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type Release struct {
	Comparison         Comparison   `json:"comparison"`
	MergedPullRequests PullRequests `json:"merged_pull_requests"`
	OpenedPullRequests PullRequests `json:"opened_pull_requests"`
	Config             Config       `json:"config"`
	Fetcher            Fetcher      `json:"-"`
}

type BuildedPullRequest struct {
	Title, Body string
}

func (rel *Release) Build(ctx context.Context) (BuildedPullRequest, error) {
	bpr := BuildedPullRequest{}

	mergedPRs, err := rel.Fetcher.MergedPullRequests(ctx)
	if err != nil {
		return bpr, err
	}
	rel.MergedPullRequests = mergedPRs

	if rel.Config.TargetLabel != "" {
		openedPRs, err := rel.Fetcher.OpenedPullRequests(ctx, rel.Config.TargetLabel)
		if err != nil {
			return bpr, err
		}
		rel.OpenedPullRequests = openedPRs
	}

	rel.Comparison = rel.Fetcher.GetComparison()

	body, err := rel.renderBody()
	if err != nil {
		return bpr, err
	}

	title, err := rel.renderTitle()
	if err != nil {
		return bpr, err
	}
	bpr.Body = body
	bpr.Title = title

	return bpr, nil
}

func (rel *Release) renderBody() (string, error) {
	filename := rel.Config.Markdown
	md, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Wrap(err, "cannot read markdown file")
	}

	filtered, err := rel.Filter(string(md))
	if err != nil {
		return "", err
	}

	return filtered, nil
}

func (rel *Release) renderTitle() (string, error) {
	filtered, err := rel.Filter(rel.Config.Title)
	if err != nil {
		return "", err
	}

	return filtered, nil
}

func (rel *Release) PullRequests() PullRequests {
	bs := append(rel.MergedPullRequests, rel.OpenedPullRequests...)
	return bs
}

func (rel *Release) Filter(s string) (string, error) {
	tmpl, err := template.New("config").Parse(s)
	if err != nil {
		return "", errors.Wrap(err, "cannot parse in filter")
	}

	wbuf := &bytes.Buffer{}
	err = tmpl.Execute(wbuf, rel)
	if err != nil {
		return "", errors.Wrap(err, "cannot execute filter")
	}

	return wbuf.String(), nil
}

func (rel *Release) Call(s string) string {
	command, err := rel.call(s)
	cmd := strings.Join(append([]string{command.Path}, command.Args...), " ")
	if err != nil {
		log.Printf("[DEBUG] fail generate command: run=%s, err=%s", cmd, err)
		return ""
	}

	out, err := command.Output()
	if err != nil {
		log.Printf("[DEBUG] fail run or read output: run=%s, err=%s", cmd, err)
		return ""
	}
	trimed := strings.TrimSpace(string(out))

	return trimed
}

func (rel *Release) call(s string) (*exec.Cmd, error) {
	calls := rel.Config.Calls
	var cmd string
	for _, call := range calls {
		if call.Name == s {
			cmd = call.Command
			break
		}
	}

	command := exec.Command("sh", "-c", cmd)
	in := &bytes.Buffer{}
	command.Stdin = in
	enc := json.NewEncoder(in)
	err := enc.Encode(rel)
	if err != nil {
		return nil, errors.Wrap(err, "fail encode Release to json")
	}

	return command, nil
}

func (rel *Release) CallIf(s string) bool {
	command, err := rel.call(s)
	cmd := strings.Join(append([]string{command.Path}, command.Args...), " ")
	if err != nil {
		log.Printf("[DEBUG] fail generate command: run=%s, err=%s", cmd, err)
		return false
	}

	out, err := command.CombinedOutput()
	if err != nil {
		log.Printf("[DEBUG] fail run command: run=%s, err=%s out=%s", cmd, err, string(out))
		return false
	}

	return true
}
