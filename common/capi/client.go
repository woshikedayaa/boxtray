package capi

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/woshikedayaa/boxtray/common"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	minResponseSize = 1 << 16
	minRequestSize
	minTimout = 1 * time.Second
)

type ClientConfig struct {
	MaxResponseSize int64
	MaxRequestSize  int64
	Timeout         time.Duration
	Secret          string
}

type Client struct {
	endpoint *url.URL

	httpClient      *http.Client
	config          *ClientConfig
	websocketClient *websocket.Dialer
}

func NewClient(endpoint string, config *ClientConfig) (*Client, error) {
	c := &Client{config: config}
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
	if endp.Scheme != "http" && endp.Scheme != "https" {
		return nil, fmt.Errorf("unexceped url scheme : %s", endp.Scheme)
	}
	if len(endp.Scheme) == 0 {
		endp.Scheme = "http"
	}
	c.endpoint = endp
	// http httpClient
	c.httpClient = &http.Client{

		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: c.config.Timeout,
		},
		Timeout: c.config.Timeout,
	}
	// ws httpClient
	c.websocketClient = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: c.config.Timeout,
	}

	return c, nil
}

func (c *Client) getRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return nil, err
	}
	if len(c.config.Secret) > 0 {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Secret))
	}

	return request, nil
}

func (c *Client) doGet(ph string, arg url.Values) ([]byte, error) {

	request, err := c.getRequest(http.MethodGet, c.newEndpoint(ph, arg).String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if contentLengthStr := response.Header.Get("Content-Length"); contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length header: %w", err)
		}

		if contentLength > c.config.MaxResponseSize {
			return nil, fmt.Errorf("response too large: %d bytes", contentLength)
		}
	}

	limitedReader := &io.LimitedReader{
		R: response.Body,
		N: c.config.MaxResponseSize + 1,
	}

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if limitedReader.N <= 0 {
		body = nil
		return nil, fmt.Errorf("response exceeded maximum size of %d bytes", c.config.MaxResponseSize)
	}

	return body, nil
}

func (c *Client) doGetStream(ctx context.Context, ph string, query url.Values) (<-chan []byte, <-chan error, error) {
	header := http.Header{}
	if len(c.config.Secret) > 0 {
		header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Secret))
	}
	header.Set("Accept", "application/json")
	header.Set("Cache-Control", "no-cache")
	//
	endpoint := c.newEndpoint(ph, query)
	switch endpoint.Scheme {
	case "http":
		endpoint.Scheme = "ws"
	case "https":
		endpoint.Scheme = "wss"
	default:
	}
	conn, resp, err := c.websocketClient.DialContext(ctx, endpoint.String(), header)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, nil, fmt.Errorf("failed to establish WebSocket connection: %w", err)
	}

	reply := make(chan []byte, 16)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(reply)
			close(errChan)
			conn.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					errChan <- err
					return
				}

				if len(message) == 0 {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case reply <- message:
				}
			}
		}
	}()

	return reply, errChan, nil
}

func (c *Client) newEndpoint(ph string, query url.Values) *url.URL {
	ne := *c.endpoint
	if query != nil && len(query) > 0 {
		ne.RawQuery = common.CombineArgs(ne.Query(), query).Encode()
	}
	ne.Path = path.Clean(path.Join(ne.EscapedPath(), ph))
	return &ne
}
