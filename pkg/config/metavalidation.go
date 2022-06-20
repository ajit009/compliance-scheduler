package config

import (
	"aws-compliance-scheduler/pkg/log"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Link struct {
	Label string
	Url   string
}

type compsMeta []struct {
	Name   string `yaml:"name"`
	Values []struct {
		Value       string `yaml:"value"`
		Description string `yaml:"description"`
		Contacts    struct {
			TeamEmail        string `yaml:"teamEmail"`
			TeamManagerEmail string `yaml:"teamManagerEmail"`
		} `yaml:"contacts"`
		Links struct {
			Chat struct {
				Label string `yaml:"label"`
				URL   string `yaml:"url"`
			} `yaml:"chat"`
			OnCall struct {
				Label string `yaml:"label"`
				URL   string `yaml:"url"`
			} `yaml:"onCall"`
			Effx struct {
				Label string `yaml:"label"`
				URL   string `yaml:"url"`
			} `yaml:"effx"`
			IssueTracker struct {
				Label string `yaml:"label"`
				URL   string `yaml:"url"`
			} `yaml:"issueTracker"`
		} `yaml:"links,omitempty"`
	} `yaml:"values,omitempty"`
	ValueRegex     []string `yaml:"valueRegex"`
	ValueMinLength int      `yaml:"valueMinLength"`
	ValueMaxLength int      `yaml:"valueMaxLength"`
}

const YamlPath string = "./ajlocalmeta/tags.yml"

func ReadTagDataFromYaml() compsMeta {
	fileFromMeta, _ := filepath.Abs(YamlPath)
	yamld, err := ioutil.ReadFile(fileFromMeta)

	if err != nil {
		log.Info(err.Error())
	}

	var meta compsMeta
	err = yaml.Unmarshal(yamld, &meta)
	if err != nil {
		log.Error(err.Error())
	}
	return meta

}
