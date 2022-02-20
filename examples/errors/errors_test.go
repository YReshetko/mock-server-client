package errors_test

import (
	"context"
	goErr "errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	msc "github.com/YReshetko/mock-server-client"
	"github.com/YReshetko/mock-server-client/examples/errors"
)

type ErrorsClientSuite struct {
	suite.Suite

	client *errors.Client
	mock   msc.MockServer
}

func (c *ErrorsClientSuite) SetupSuite() {
	c.client = errors.NewClient("http://localhost:1080")
	c.mock = msc.NewMockServer(msc.Config{
		Host:    "localhost",
		Port:    1080,
		Verbose: true,
	})
}

func (c *ErrorsClientSuite) TestTimeout() {
	e := c.mock.On(http.MethodGet, "/some/endpoint").
		Name("Timeout endpoint submission").
		DefaultResponse(
			msc.WithStatusCode(http.StatusOK),
			msc.WithDelay(time.Second*2),
		).
		//NumCalls(2).
		NumCalls(1).
		AssertionAtCall(0, msc.NewAssertion().WithNoBody())

	c.Require().NoError(c.mock.Setup(context.Background(), e))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.client.Do(ctx)
	c.Require().Error(err)
	c.Equal(errors.TimeoutError, err)

	c.Require().NoError(c.mock.Verify(context.Background(), c.T()))
}

func (c *ErrorsClientSuite) TestHTTPError() {
	e := c.mock.On(http.MethodGet, "/some/endpoint").
		Name("Drop connection").
		DefaultResponse(
			msc.WithStatusCode(http.StatusOK),
			msc.WithDelay(time.Millisecond*500),
			msc.WithDropConnection(),
			msc.WithErrorBytes([]byte("eQqmdjEEoaXnCvcK6lOAIZeU+Pn+womxmg==")),
		).
		NumCalls(1).
		AssertionAtCall(0, msc.NewAssertion().WithNoBody())

	c.Require().NoError(c.mock.Setup(context.Background(), e))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := c.client.Do(ctx)
	c.Require().Error(err)
	c.True(goErr.Is(err, errors.UnexpectedError))

	c.Require().NoError(c.mock.Verify(context.Background(), c.T()))
}

func (c *ErrorsClientSuite) SetupTest() {
	c.Require().NoError(c.mock.Reset(context.Background()))
}

func TestErrorsClientSuite(t *testing.T) {
	suite.Run(t, &ErrorsClientSuite{})
}
