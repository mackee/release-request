package main

import (
	"testing"
)

func TestReleaser__Filter(t *testing.T) {
	r := &Release{
		MergedPullRequests: PullRequests{
			{Title: "topic/hoge"},
			{Title: "topic/fuga"},
		},
	}

	in := "{{ .PullRequests.Titles }}"

	out, err := r.Filter(in)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if out != "topic/hoge topic/fuga" {
		t.Errorf("not match result: %s", out)
	}
}
