package capi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	minResponseSize = 4096
	minRequestSize
	minTimout = 50 * time.Millisecond
)

type ClientConfig struct {
	MaxResponseSize int64
	MaxRequestSize  int64
	Timeout         time.Duration
	Secret          string
}

type Client struct {
	ctx      context.Context
	endpoint *url.URL

	client *http.Client
	config *ClientConfig
}

func NewClient(ctx context.Context, endpoint string, config *ClientConfig) (*Client, error) {
	c := &Client{config: config, ctx: ctx}
	if c.ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if c.config == nil {
		c.config = &ClientConfig{}
	}
	c.config.Timeout = max(minTimout, c.config.Timeout)
	c.config.MaxResponseSize = max(minResponseSize, c.config.MaxResponseSize)
	c.config.MaxRequestSize = max(minRequestSize, c.config.MaxRequestSize)

	endp, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	c.endpoint = endp
	c.client = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: c.config.Timeout,
		},
		Timeout: c.config.Timeout,
	}

	return c, nil
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

func (c *Client) getRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(c.ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if len(c.config.Secret) > 0 {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Secret))
	}

	return request, nil
}

func (c *Client) doGet(ph string, arg url.Values) ([]byte, error) {
	ne := *c.endpoint
	if arg != nil && len(arg) > 0 {
		// shit code
		for k, v := range ne.Query() {
			for _, vv := range v {
				arg.Add(k, vv)
			}
		}
		ne.RawQuery = arg.Encode()
	}
	ne.Path = strings.TrimRight(strings.TrimLeft(path.Join(ne.EscapedPath(), ph), "/"), "/")

	request, err := c.getRequest(http.MethodGet, ne.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 检查状态码
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	// 处理 Content-Length 头
	if contentLengthStr := response.Header.Get("Content-Length"); contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length header: %w", err)
		}

		if contentLength > c.config.MaxResponseSize {
			return nil, fmt.Errorf("response too large: %d bytes", contentLength)
		}
	}

	// 即使没有 Content-Length 头，也要限制读取大小
	limitedReader := &io.LimitedReader{
		R: response.Body,
		N: c.config.MaxResponseSize + 1, // +1 用于检测是否超过最大大小
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// 如果 limitedReader.N 为 0，说明达到了限制大小
	if limitedReader.N <= 0 {
		body = nil
		return nil, fmt.Errorf("response exceeded maximum size of %d bytes", c.config.MaxResponseSize)
	}

	return body, nil
}
