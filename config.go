package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kayac/go-config"
	"github.com/pkg/errors"
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
	Owner            string   `yaml:"owner"`
	Repo             string   `yaml:"repo"`
	Head             string   `yaml:"head"`
	LogContainsBases []string `yaml:"log_contains_bases"`
	Base             string   `yaml:"base"`
}

func (r Repo) JoinedHead() string {
	if strings.Contains(r.Head, ":") {
		return r.Head
	}
	return strings.Join([]string{r.Owner, r.Head}, ":")
}

func (r Repo) ContainsBase(base string) bool {
	bases := append(r.LogContainsBases, r.Head)
	for _, _base := range bases {
		if _base == base {
			return true
		}
	}
	return false
}

func NewConfig(r io.Reader) (Config, error) {
	var c Config

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return c, errors.Wrap(err, "cannot read config reader")
	}

	loader := config.New()
	loader.Delims("%%", "%%")
	err = loader.LoadWithEnvBytes(&c, bs)
	if err != nil {
		return c, errors.Wrap(err, "cannot unmarshal config")
	}

	if c.Token == "" {
		c.Token = os.Getenv("GITHUB_TOKEN")
	}

	return c, nil
}
