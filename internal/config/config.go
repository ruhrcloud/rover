package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type FromConf struct {
	Host    string `yaml:"host"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
	Mailbox string `yaml:"mailbox"`
}

type ToConf struct {
	BaseURL string `yaml:"base_url"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
	Auth    string `yaml:"auth"`
	Token   string `yaml:"token"`
}

type FilterConf struct {
	Recipients []string `yaml:"recipients"`
	Seen       *bool    `yaml:"seen"`
	Extensions []string `yaml:"extensions"`
}

type Task struct {
	Name     string     `yaml:"name"`
	From     FromConf   `yaml:"from"`
	To       ToConf     `yaml:"to"`
	Path     string     `yaml:"path"`
	Tags     []string   `yaml:"tags"`
	Filter   FilterConf `yaml:"filter"`
	Interval string     `yaml:"interval"`
	Format   string     `yaml:"format"`
	MarkSeen bool       `yaml:"mark_seen"`
}

type Config struct {
	Tasks []Task `yaml:"tasks"`
}

func Load(p string) (*Config, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	for i := range c.Tasks {
		if err := validateTask(c.Tasks[i]); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func validateTask(t Task) error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("task.name required")
	}

	if t.From.User == "" || t.From.Pass == "" {
		return errors.New("from.user/from.pass required")
	}
	if t.From.Mailbox == "" {
		return errors.New("from.mailbox required")
	}
	if t.To.BaseURL == "" {
		return errors.New("to.base_url required")
	}
	switch strings.ToLower(strings.TrimSpace(t.To.Auth)) {
	case "", "basic":
		if t.To.User == "" || t.To.Pass == "" {
			return errors.New("to.user/to.pass required for basic auth")
		}
	case "bearer":
		if t.To.Token == "" {
			return errors.New("to.token required for bearer auth")
		}
	default:
		return errors.New("to.auth must be basic or bearer")
	}
	if _, err := time.ParseDuration(t.Interval); err != nil {
		return errors.New("interval must be a valid duration")
	}
	if strings.TrimSpace(t.Format) == "" {
		return errors.New("format template required")
	}
	return nil
}

