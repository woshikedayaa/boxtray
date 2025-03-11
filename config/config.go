package config

import (
	"encoding/json"
	"io"
	"net/url"
)

type ControlConfig struct {
	Start []string `json:"start"`
	Stop  []string `json:"stop"`

	Update []string `json:"update"`
}

type ApiConfig struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Secret string `json:"secret"`

	Label string `json:"label"`

	Control ControlConfig `json:"control"`
}

func (a ApiConfig) DisplayName() string {
	if a.Label != "" {
		return a.Label
	}
	return a.Host
}

func (a ApiConfig) Endpoint() string {
	if a.Scheme == "" || a.Scheme != "http" && a.Scheme != "https" {
		a.Scheme = "http"
	}
	u := url.URL{}
	u.Scheme = a.Scheme
	u.Host = a.Host
	u.Path = a.Path

	return u.String()
}

type LogConfig struct {
	Level   string `json:"level"`
	Disable bool   `json:"disable"`
}
type Config struct {
	Api ApiConfig `json:"api"`
	Log LogConfig `json:"log"`
}

func New(reader io.Reader) (*Config, error) {
	cfg := &Config{}
	err := json.NewDecoder(reader).Decode(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
