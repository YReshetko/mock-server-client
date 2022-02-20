package regexp

import (
	"github.com/pkg/errors"
	"net/http"
)

const path = "/some/endpoint"

type Client struct {
	client *http.Client
	host   string
}

func NewClient(host string) *Client {
	return &Client{
		client: http.DefaultClient,
		host:   host,
	}
}

func (c *Client) Do(headers map[string][]string, queryLine string) error {
	rq, err := http.NewRequest(http.MethodGet, c.host+path+"?"+queryLine, nil)
	if err != nil {
		return errors.Wrap(err, "unable to prepare HTTP request")
	}

	for k, strings := range headers {
		for _, s := range strings {
			rq.Header.Add(k, s)
		}
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		return errors.Wrap(err, "unable to submit form")
	}
	defer rs.Body.Close()

	return nil
}

func (c *Client) BasicAuth(username, password string) error {
	rq, err := http.NewRequest(http.MethodPost, c.host+path, nil)
	if err != nil {
		return errors.Wrap(err, "unable to prepare HTTP request")
	}

	rq.SetBasicAuth(username, password)

	rs, err := c.client.Do(rq)
	if err != nil {
		return errors.Wrap(err, "unable to submit form")
	}
	defer rs.Body.Close()

	return nil
}
