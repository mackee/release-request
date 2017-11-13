package main

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Title       string   `yaml:"title"`
	Repo        Repo     `yaml:"repo"`
	Labels      []string `yaml:"labels"`
	TargetLabel string   `yaml:"target_label"`
	Calls       []call   `yaml:"calls"`
	Token       string   `yaml:"token"`
	Checkout    bool     `yaml:"checkout"`
	Markdown    string   `yaml:"markdown"`
}

type call struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type Repo struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
	Head  string `yaml:"head"`
	Base  string `yaml:"base"`
}

func (r Repo) JoinedHead() string {
	if strings.Contains(r.Head, ":") {
		return r.Head
	}
	return strings.Join([]string{r.Owner, r.Head}, ":")
}

func NewConfig(r io.Reader) (Config, error) {
	var c Config

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return c, errors.Wrap(err, "cannot read config reader")
	}

	err = yaml.Unmarshal(bs, &c)
	if err != nil {
		return c, errors.Wrap(err, "cannot unmarshal config")
	}

	return c, nil
}
