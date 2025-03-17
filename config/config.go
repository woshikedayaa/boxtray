package config

import (
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

	Control ControlConfig `json:"control"`
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

type BoxConfig struct {
	UrlTest  string `json:"url_test"`
	MaxDelay uint16 `json:"max_delay"`
}
type Config struct {
	Api ApiConfig `json:"api"`
	Log LogConfig `json:"log"`
	Box BoxConfig `json:"box"`
}
