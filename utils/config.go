package utils

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Targets []Targets `yaml:"targets"`
}

type Targets struct {
	IpAddress string `yaml:"ipAddress"`
	Userid    string `yaml:"userid"`
	Password  string `yaml:"password"`
}

func GetConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg = new(Config)

	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
