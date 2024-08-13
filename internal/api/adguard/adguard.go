package adguard

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/condrove10/echo/pkg/retryablehttp"
	"github.com/go-playground/validator/v10"
)

type record struct {
	Name    string `json:"domain"`
	Content string `json:"answer"`
}

type Client struct {
	Url         string          `validate:"required,http_url"`
	authHeader  string          `validate:"omitempty"`
	context     context.Context `validate:"required"`
	Retries     uint16          `validate:"required,gt=0,lte=65535"`
	RetryDelay  time.Duration   `validate:"required"`
	InsecureTls bool            `validate:"required"`
}

func New(options ...func(*Client)) (*Client, error) {
	client := Client{
		context: context.Background(),
	}

	for _, option := range options {
		option(&client)
	}

	v := validator.New()

	if err := v.Struct(&client); err != nil {
		return nil, err
	}

	return &client, nil
}

func WithUrl(value string) func(*Client) {
	return func(c *Client) {
		c.Url = value
	}
}

func WithContext(value context.Context) func(*Client) {
	return func(c *Client) {
		c.context = value
	}
}

func WithAuthHeader(value string) func(*Client) {
	return func(c *Client) {
		c.authHeader = value
	}
}

func WithRetries(value uint16) func(*Client) {
	return func(c *Client) {
		c.Retries = value
	}
}

func WithRetryDelay(value time.Duration) func(*Client) {
	return func(c *Client) {
		c.RetryDelay = value
	}
}

func WithInsecureTls(value bool) func(*Client) {
	return func(c *Client) {
		c.InsecureTls = value
	}
}

func (c *Client) ListRecords() (map[string]string, error) {
	client, err := retryablehttp.New(
		retryablehttp.WithHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: c.InsecureTls,
				},
			},
		}),
		retryablehttp.WithUrl(c.Url+"/control/rewrite/list"),
		retryablehttp.AddHeader("Authorization", c.authHeader),
		retryablehttp.WithMethod(retryablehttp.MethodGet),
		retryablehttp.WithMaxRetries(c.Retries),
		retryablehttp.WithRetryDelay(c.RetryDelay),
		retryablehttp.WithRetryCondition(func(resp *http.Response, err error) bool {
			return resp.StatusCode != 200 && err != nil
		}),
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do()
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var records []*record
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, err
	}

	list := map[string]string{}
	for _, record := range records {
		if list[record.Name] == "" {
			list[record.Name] = record.Content
		} else {
			c.RemoveRecord(record.Name, record.Content)
		}
	}

	return list, nil
}

func (c *Client) checkRecordStatus(name, content string) (bool, error) {
	records, err := c.ListRecords()
	if err != nil {
		return false, err
	}

	if records[name] == content {
		return true, nil
	}

	return false, nil
}

func (c *Client) getNameContent(name string) (string, error) {
	records, err := c.ListRecords()
	if err != nil {
		return "", err
	}

	return records[name], nil
}

func (c *Client) AppendRecord(name string, content string) error {
	status, err := c.checkRecordStatus(name, content)
	if err != nil {
		return err
	}

	if status {
		return nil
	}

	remoteContent, err := c.getNameContent(name)
	if err != nil {
		return err
	}

	if len(remoteContent) > 0 {
		if err := c.RemoveRecord(name, remoteContent); err != nil {
			return err
		}
	}

	record := record{
		Name:    name,
		Content: content,
	}

	serialization, err := json.Marshal(&record)
	if err != nil {
		return err
	}

	client, err := retryablehttp.New(
		retryablehttp.WithHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: c.InsecureTls,
				},
			},
		}),
		retryablehttp.WithUrl(c.Url+"/control/rewrite/add"),
		retryablehttp.AddHeader("Content-Type", "application/json"),
		retryablehttp.AddHeader("Authorization", c.authHeader),
		retryablehttp.WithMethod(retryablehttp.MethodPost),
		retryablehttp.WithBody(serialization),
		retryablehttp.WithMaxRetries(c.Retries),
		retryablehttp.WithRetryDelay(c.RetryDelay),
		retryablehttp.WithRetryCondition(func(resp *http.Response, err error) bool {
			return resp.StatusCode != 200 && err != nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = client.Do()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveRecord(name string, content string) error {
	record := record{
		Name:    name,
		Content: content,
	}

	serialization, err := json.Marshal(&record)
	if err != nil {
		return err
	}

	client, err := retryablehttp.New(
		retryablehttp.WithHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: c.InsecureTls,
				},
			},
		}),
		retryablehttp.WithUrl(c.Url+"/control/rewrite/delete"),
		retryablehttp.AddHeader("Content-Type", "application/json"),
		retryablehttp.AddHeader("Authorization", c.authHeader),
		retryablehttp.WithMethod(retryablehttp.MethodPost),
		retryablehttp.WithBody(serialization),
		retryablehttp.WithMaxRetries(c.Retries),
		retryablehttp.WithRetryDelay(c.RetryDelay),
		retryablehttp.WithRetryCondition(func(resp *http.Response, err error) bool {
			return resp.StatusCode != 200 && err != nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = client.Do()
	if err != nil {
		return err
	}

	return nil
}
