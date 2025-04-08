package config

import (
	"errors"
	"net/url"
)

type Mode string

const (
	PollingMode Mode = "polling"
	WebhookMode Mode = "webhook"
)

type Bot struct {
	Token   string `yaml:"token"`
	Debug   bool   `yaml:"debug"`
	Mode    Mode   `yaml:"mode"`
	Polling `yaml:"polling"`
	Webhook `yaml:"webhook"`
}

func (b *Bot) validate() error {
	if b.Token == "" {
		return errors.New("bot token is required")
	}
	switch b.Mode {
	case PollingMode:
		if err := b.Polling.validate(); err != nil {
			return err
		}
	case WebhookMode:
		if err := b.Webhook.validate(); err != nil {
			return err
		}
	default:
		return errors.New("invalid mode")
	}
	return nil
}

type Polling struct {
	Offset  int `yaml:"offset"`
	Limit   int `yaml:"limit"`
	Timeout int `yaml:"timeout"`
}

func (p *Polling) validate() error {
	if p.Limit < 1 || p.Limit > 100 {
		return errors.New("polling limit must be between 1 and 100")
	}
	if p.Timeout < 0 {
		return errors.New("polling timeout must be non-negative")
	}
	return nil
}

type Webhook struct {
	URL                string `yaml:"url"`
	Port               int    `yaml:"port"`
	MaxConn            int    `yaml:"max_conn"`
	DropPendingUpdates bool   `yaml:"drop_pending_updates"`
	SSL                struct {
		Enabled bool   `yaml:"enabled"`
		Cert    string `yaml:"cert"`
		Key     string `yaml:"key"`
	} `yaml:"ssl"`
}

func (w *Webhook) validate() error {
	if w.URL == "" {
		return errors.New("webhook url is required")
	}
	if _, err := url.ParseRequestURI(w.URL); err != nil {
		return err
	}
	if w.Port < 1 || w.Port > 65535 {
		return errors.New("webhook port must be between 1 and 65535")
	}
	if w.MaxConn < 1 {
		return errors.New("webhook max connections must be non-negative")
	}
	if w.SSL.Enabled {
		if w.SSL.Cert == "" {
			return errors.New("webhook .ssl cert is required")
		}
		if w.SSL.Key == "" {
			return errors.New("webhook .ssl key is required")
		}
	}
	return nil
}
