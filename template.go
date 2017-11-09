package main

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

var (
	templateFiles      = []string{"config.yml", "release.md"}
	templateConfigFile = "config.yml"
)

//go:generate go-assets-builder --output=template.gen.go --strip-prefix=/_example/ _example

func GenerateTemplates(configfile string) error {
	_, err := os.Stat(configfile)
	if !os.IsNotExist(err) {
		return errors.Wrap(err, "configfile is exists")
	}

	for _, filename := range templateFiles {
		f, err := Assets.Open(filename)
		if err != nil {
			return errors.Wrapf(err, "fail open from assets: %s", filename)
		}
		defer f.Close()

		output := filename
		if filename == templateConfigFile {
			output = configfile
		}
		dest, err := os.Create(output)
		if err != nil {
			return errors.Wrapf(err, "fail create new file: %s", output)
		}
		defer dest.Close()

		_, err = io.Copy(dest, f)
		if err != nil {
			return errors.Wrapf(err, "fail copy to new file from assets: %s", output)
		}
	}

	return nil
}
