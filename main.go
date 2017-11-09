package main

import (
	"flag"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

func main() {
	var configfile, loglevel string
	var isDryRun, isInit bool
	flag.StringVar(&configfile, "config", "config.yml", "specify config file")
	flag.StringVar(&loglevel, "loglevel", "WARN", "log level. ex.ERROR,WARN,DEBUG")
	flag.BoolVar(&isDryRun, "dry-run", false, "no create/update pull request")
	flag.BoolVar(&isInit, "init", false, "generate sample of configuration files")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(loglevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	if isInit {
		err := GenerateTemplates(configfile)
		if err != nil {
			log.Fatalf("[ERROR] cannot generate samples: %s", err)
		}
		return
	}

	f, err := os.Open(configfile)
	if os.IsNotExist(err) {
		log.Fatalf("[WARN] configfile is not exists. If you want sample configuration file, execute `release-request -init`")
	}
	if err != nil {
		log.Fatalf("[ERROR] cannot read file: %s", err)
	}
	defer f.Close()

	c, err := NewConfig(f)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	runner := NewRunner(c)
	if err := runner.Run(isDryRun); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
}
