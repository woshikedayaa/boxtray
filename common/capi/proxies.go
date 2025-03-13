package capi

import (
	"encoding/json"
	"fmt"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"net/http"
	"path"
	"strings"
	"time"
)

type Proxies struct {
	Proxies *orderedmap.OrderedMap[string, *Proxy] `json:"proxies"`
}

type History struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
}
type Proxy struct {
	Type    string    `json:"type"`
	Name    string    `json:"name"`
	UDP     bool      `json:"udp"`
	History []History `json:"history"`
	//
	Now string   `json:"now"`
	All []string `json:"all"`
}

func (c *Client) GetProxies() (*Proxies, error) {
	bs, err := c.doGet("/proxies", nil)
	if err != nil {
		return nil, err
	}
	p := &Proxies{orderedmap.New[string, *Proxy]()}
	err = json.Unmarshal(bs, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Client) SwitchProxy(selector string, target string) error {
	req, err := c.getRequest(http.MethodPut, c.newEndpoint(path.Join("proxies", selector), nil).String(), strings.NewReader(fmt.Sprintf("{\"name\":\"%s\"}", target)))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("unexcepted status code : %d", resp.StatusCode)
}
