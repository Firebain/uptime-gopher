package main

import (
	"time"
)

type CheckConfig struct {
	Key      string            `yaml:"key"`
	Interval time.Duration     `yaml:"interval"`
	Args     map[string]string `yaml:",inline"`
}

type Domain struct {
	Domain   string        `yaml:"domain"`
	Interval time.Duration `yaml:"interval"`
	Checks   []CheckConfig `yaml:"checks"`
}

type Config struct {
	Domains []Domain `yaml:"domains"`
}
