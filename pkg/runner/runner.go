package runner

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/fgimenez/cihealth/pkg/gh"
	"github.com/fgimenez/cihealth/pkg/stats"
	"github.com/fgimenez/cihealth/pkg/types"
)

func Run(o *types.Options) (string, error) {
	if o.LogLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if o.TokenPath == "" {
		return "", fmt.Errorf("You need to specify the GitHub token path with --gh-token")
	}

	ghClient, err := gh.NewClient(o.TokenPath, o.Source, o.DataDays)
	if err != nil {
		return "", err
	}

	results, err := stats.Run(ghClient)
	if err != nil {
		return "", err
	}

	d, err := yaml.Marshal(&results)
	if err != nil {
		return "", err
	}

	if o.Path == "" {
		file, err := ioutil.TempFile("", "cihealth")
		if err != nil {
			log.Fatal(err)
		}
		o.Path = file.Name()
	}

	log.Printf("Writing output file %s", o.Path)
	err = ioutil.WriteFile(o.Path, d, 0644)
	if err != nil {
		return "", err
	}

	return o.Path, nil
}

func setLogLevel(logLevel string) {
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
