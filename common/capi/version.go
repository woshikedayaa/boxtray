package capi

import "encoding/json"

type Version struct {
	Meta    bool   `json:"meta"`
	Premium bool   `json:"premium"`
	Version string `json:"version"`
}

func (c *Client) GetVersion() (Version, error) {
	bs, err := c.doGet("/version", nil)
	if err != nil {
		return Version{}, err
	}
	v := Version{}
	err = json.Unmarshal(bs, &v)
	return v, err
}
