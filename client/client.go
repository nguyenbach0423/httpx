package client

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/nguyenbach0423/httpx/client/request"
	"github.com/nguyenbach0423/httpx/client/response"
)

type Client struct {
	httpclient  *http.Client
	retryConfig *RetryConfig
}

func New(opts ...func(*Client)) *Client {
	c := &Client{
		httpclient: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		c.httpclient.Timeout = timeout
	}
}

type RetryConfig struct {
	MaxRetries int
	Backoff    time.Duration
	MaxBackoff time.Duration
}

func WithRetryConfig(retryConfig *RetryConfig) func(*Client) {
	return func(c *Client) {
		c.retryConfig = retryConfig
	}
}

func (c *Client) Do(req *request.Request) (*response.Response, error) {
	var resp *response.Response

	reqBody := req.Body

	var maxRetries int
	var backoff time.Duration
	var maxBackoff time.Duration

	if c.retryConfig != nil {
		maxRetries = c.retryConfig.MaxRetries
		backoff = c.retryConfig.Backoff
		maxBackoff = c.retryConfig.MaxBackoff
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var reader io.Reader
		if len(reqBody) > 0 {
			reader = bytes.NewReader(reqBody)
		}

		var err error

		var httpReq *http.Request
		httpReq, err = http.NewRequest(req.Method, buildURL(req), reader)
		if err != nil {
			return nil, err
		}

		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}

		var httpResp *http.Response
		httpResp, err = c.httpclient.Do(httpReq)
		if err != nil {
			return nil, err
		}

		var respBody []byte
		respBody, err = io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()
		if err != nil {
			return nil, err
		}

		resp = &response.Response{
			Status: httpResp.StatusCode,
			Body:   respBody,
		}

		if resp.Status == http.StatusOK {
			return resp, nil
		}

		if resp.Status >= 500 && resp.Status <= 599 {
			if attempt < maxRetries {
				sleepTime := backoff * (1 << attempt)
				if sleepTime > maxBackoff {
					sleepTime = maxBackoff
				}

				if sleepTime > 0 {
					sleepTime = time.Duration(rand.Int63n(int64(sleepTime)))

					time.Sleep(sleepTime)
				}
			}
			continue
		}
	}

	return resp, nil
}

func buildURL(r *request.Request) string {
	queryParams := r.QueryParams
	if queryParams == nil || len(queryParams) == 0 {
		return r.URL
	}

	values := url.Values{}

	for k, v := range queryParams {
		for _, i := range v {
			values.Add(k, i)
		}
	}

	return r.URL + "?" + values.Encode()
}
