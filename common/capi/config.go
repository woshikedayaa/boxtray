package capi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	Port        int    `json:"port"`
	SocksPort   int    `json:"socks-port"`
	RedirPort   int    `json:"redir-port"`
	TProxyPort  int    `json:"tproxy-port"`
	MixedPort   int    `json:"mixed-port"`
	AllowLan    bool   `json:"allow-lan"`
	BindAddress string `json:"bind-address"`
	Mode        string `json:"mode"`
	// sing-box added
	ModeList []string       `json:"mode-list"`
	LogLevel string         `json:"log-level"`
	IPv6     bool           `json:"ipv6"`
	Tun      map[string]any `json:"tun"`
}

func (c *Client) GetConfig() (*Config, error) {
	bs, err := c.doGet("/config", nil)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = json.Unmarshal(bs, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Client) SetMode(mode string) error {
	if len(mode) == 0 {
		return fmt.Errorf("mode str can not be empty")
	}
	req, err := http.NewRequest(http.MethodPatch, c.newEndpoint("/config", nil).String(), strings.NewReader(fmt.Sprintf("{\"mode\":\"%s\"}", mode)))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
