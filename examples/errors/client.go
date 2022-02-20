package errors

import (
	"context"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const path = "/some/endpoint"

var (
	TimeoutError    = errors.New("timeout error")
	UnexpectedError = errors.New("unexpected error")
)

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

func (c *Client) Do(ctx context.Context) error {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.host+path, nil)
	if err != nil {
		return errors.Wrap(err, "unable to prepare HTTP request")
	}
	rs, err := c.client.Do(rq)
	if rs != nil && rs.Body != nil {
		defer rs.Body.Close()
	}

	if os.IsTimeout(err) {
		return TimeoutError
	}
	if err != nil {
		return errors.Wrapf(UnexpectedError, "; reason: %s", err)
	}

	return nil
}
