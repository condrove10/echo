package retryablehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type Method string

const (
	MethodGet    Method = "GET"
	MethodHead   Method = "HEAD"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodPatch  Method = "PATCH"
	MethodDelete Method = "DELETE"
	// MethodConnect Method = "CONNECT"
	// MethodOptions Method = "OPTIONS"
	// MethodTrace   Method = "TRACE"
)

type Client struct {
	Url            string                                    `validate:"required,http_url"`
	Method         Method                                    `validate:"required"`
	Body           []byte                                    `validate:"required"`
	Header         http.Header                               `validate:"required"`
	HttpClient     *http.Client                              `validate:"required"`
	context        context.Context                           `validate:"required"`
	maxRetries     uint16                                    `validate:"required,gt=0,lte=65535"`
	retryDelay     time.Duration                             `validate:"required"`
	retryCondition func(resp *http.Response, err error) bool `validate:"required"`
}

func New(options ...func(*Client)) (*Client, error) {
	// Initialize a new retryable HTTP client.
	client := Client{
		context: context.Background(),
		Header:  make(http.Header),
		Body:    make([]byte, 0),
	}

	for _, option := range options {
		option(&client)
	}

	// Initialize a new validator.
	v := validator.New()

	// Validate the Server struct.
	if err := v.Struct(&client); err != nil {
		return nil, err
	}

	return &client, nil
}

func WithHttpClient(client *http.Client) func(*Client) {
	return func(c *Client) {
		c.HttpClient = client
	}
}

func WithUrl(url string) func(*Client) {
	return func(c *Client) {
		c.Url = url
	}
}

func WithMethod(method Method) func(*Client) {
	return func(c *Client) {
		c.Method = method
	}
}

func WithBody(body []byte) func(*Client) {
	return func(c *Client) {
		c.Body = body
	}
}

func WithHeader(header http.Header) func(*Client) {
	return func(c *Client) {
		c.Header = header.Clone()
	}
}

func AddHeader(key string, value string) func(*Client) {
	return func(c *Client) {
		if c.Header == nil {
			c.Header = make(http.Header)
		}
		c.Header.Add(key, value)
	}
}

func WithContext(ctx context.Context) func(*Client) {
	return func(c *Client) {
		c.context = ctx
	}
}

func WithMaxRetries(maxRetries uint16) func(*Client) {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

func WithRetryDelay(retryDelay time.Duration) func(*Client) {
	return func(c *Client) {
		c.retryDelay = retryDelay
	}
}

func WithRetryCondition(retryCondition func(resp *http.Response, err error) bool) func(*Client) {
	return func(c *Client) {
		c.retryCondition = retryCondition
	}
}

func (c *Client) retry(fn func() (*http.Response, error)) (*http.Response, error) {
	var (
		resp            = &http.Response{}
		err      error  = nil
		retry    bool   = false
		retryErr string = ""
	)

	for i := uint16(0); i < c.maxRetries; i += 1 {
		if c.context.Err() != nil {
			err := fmt.Errorf("retryable http call context closed; %v", c.context.Err())
			return nil, err
		}

		if retry {
			time.Sleep(c.retryDelay)
		}

		resp, err = fn()

		retry = c.retryCondition(resp, err)

		if err == nil && !retry {
			return resp, nil
		}

		retryErr += fmt.Sprintf("\n\t\ttry %d: %v; retry condition status: %v", i+1, err, retry)
	}

	return nil, fmt.Errorf("retryable http client max retries exceeded; %s", retryErr)
}

func (c *Client) Do() (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.context, string(c.Method), c.Url, bytes.NewReader(c.Body))
	if err != nil {
		return nil, err
	}

	return c.retry(func() (*http.Response, error) {
		req.Body = io.NopCloser(bytes.NewReader(c.Body))

		req.Header = c.Header.Clone()

		return c.HttpClient.Do(req)
	})
}
