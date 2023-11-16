package main

import (
	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/opensourceways/server-common-lib/utils"
)

type Config struct {
	UserAgent   string       `json:"user_agent"    required:"true"`
	GroupName   string       `json:"group_name"    required:"true"`
	Token       string       `json:"token"         required:"true"`
	CIRobotName string       `json:"ci_robot_name" required:"true"`
	Topics      Topics       `json:"topics"`
	Kafka       kafka.Config `json:"kafka"`
	Repository  Repository   `json:"repository"`
}

type Topics struct {
	SoftwarePkgHookEvent string `json:"software_pkg_hook_event" required:"true"`
	SoftwarePkgCIChecked string `json:"software_pkg_ci_checked" required:"true"`
}

type Repository struct {
	Org  string `json:"org"`
	Repo string `json:"repo"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := utils.LoadFromYaml(path, cfg); err != nil {
		return nil, err
	}

	cfg.SetDefault()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.Kafka,
		&cfg.Repository,
	}
}

func (cfg *Config) SetDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (cfg *Config) Validate() error {
	if _, err := utils.BuildRequestBody(cfg, ""); err != nil {
		return err
	}

	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) SetDefault() {
	if r.Org == "" {
		r.Org = "src-openeuler"
	}

	if r.Repo == "" {
		r.Repo = "software-package-server"
	}
}
