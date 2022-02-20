package form

import (
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

const path = "/form/submit"

type Client struct {
	client *http.Client
	host   string
}

func NewFormClient(host string) *Client {
	return &Client{
		client: http.DefaultClient,
		host:   host,
	}
}

func (c *Client) Submit(values map[string]string) error {
	val := url.Values{}
	for k, v := range values {
		val[k] = []string{v}
	}
	rs, err := c.client.PostForm(c.host+path, val)
	if err != nil {
		return errors.Wrap(err, "unable to submit form")
	}
	defer rs.Body.Close()

	return nil
}
