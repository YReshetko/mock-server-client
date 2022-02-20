package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var _ Client = (*client)(nil)

type Client interface {
	Expectation(context.Context, Expectation) error
	Verify(context.Context, Verify) error
	VerifySequence(context.Context, VerifySequence) error

	Clear(context.Context, ClearRequest) error
	Reset(context.Context) error
	Retrieve(context.Context, RetrieveRequest) (RetrieveResponse, error)
}

type client struct {
	client       *http.Client
	basePath     string
	verboseError bool
}

func NewClient(host string, port int, verbose bool) *client {
	return &client{
		client:       http.DefaultClient,
		basePath:     fmt.Sprintf("http://%s:%d", host, port),
		verboseError: verbose,
	}
}

func (c *client) do(ctx context.Context, uri string, rq, rs interface{}) error {
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	reader := &bytes.Reader{}
	if rq != nil {
		data, err := json.Marshal(rq)
		if err != nil {
			return errors.Wrap(err, "unable to marshal request")
		}
		reader = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, c.basePath+uri, reader)
	if err != nil {
		return errors.Wrap(err, "unable to prepare http request")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "unable to call mockserver")
	}

	var body []byte
	if response.Body != nil {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "unable to read response body")
		}
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		var msg string
		if c.verboseError {
			msg = fmt.Sprintf("unexpected http status %d instead of 2xx; response: %s", response.StatusCode, string(body))
		} else {
			msg = fmt.Sprintf("unexpected http status %d instead of 2xx", response.StatusCode)
		}
		return errors.New(msg)
	}

	if rs == nil {
		return nil
	}

	return errors.Wrap(json.Unmarshal(body, rs), "unable to unmarshal response")
}

const expectationURI = "/expectation"

func (c *client) Expectation(ctx context.Context, request Expectation) error {
	return errors.Wrap(
		c.do(ctx, expectationURI, request, nil),
		"unable to setup expectation",
	)
}

const verifyURI = "/verify"

func (c *client) Verify(ctx context.Context, request Verify) error {
	return errors.Wrap(
		c.do(ctx, verifyURI, request, nil),
		"unable to verify expectation",
	)
}

const verifySequenceURI = "/verifySequence"

func (c *client) VerifySequence(ctx context.Context, request VerifySequence) error {
	return errors.Wrap(
		c.do(ctx, verifySequenceURI, request, nil),
		"unable to verify sequence expectations",
	)
}

const clearURI = "/clear"

func (c *client) Clear(ctx context.Context, request ClearRequest) error {
	return errors.Wrap(
		c.do(ctx, clearURI, request, nil),
		"unable to clear expectation",
	)
}

const resetURI = "/reset"

func (c *client) Reset(ctx context.Context) error {
	return errors.Wrap(
		c.do(ctx, resetURI, nil, nil),
		"unable to reset all expectations",
	)
}

const retrieveURI = "/retrieve"

func (c *client) Retrieve(ctx context.Context, request RetrieveRequest) (RetrieveResponse, error) {
	rs := RetrieveResponse{}
	err := errors.Wrap(
		c.do(ctx, retrieveURI, request, &rs),
		"unable to reset all expectations",
	)
	return rs, err
}
