package capi

import (
	"encoding/json"
	"path"
	"strconv"
)

type Delay struct {
	Delay uint16 `json:"delay"`
}

func (c *Client) GetDelay(target string, url string, timeout int) (Delay, error) {
	if url == "" {
		url = "https://google.com/generate_204"
	}
	if timeout <= 0 {
		timeout = 500
	}
	bs, err := c.doGet(path.Join("proxies", target, "delay"), map[string][]string{
		"url":     []string{url},
		"timeout": []string{strconv.FormatInt(int64(timeout), 10)},
	})
	if err != nil {
		return Delay{}, err
	}
	d := Delay{}
	err = json.Unmarshal(bs, &d)
	if err != nil {
		return Delay{}, err
	}
	return d, nil
}
