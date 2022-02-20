package form_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	msc "github.com/YReshetko/mock-server-client"
	"github.com/YReshetko/mock-server-client/examples/form"
)

type FormClientSuite struct {
	suite.Suite

	client *form.Client
	mock   msc.MockServer
}

func (c *FormClientSuite) SetupSuite() {
	c.client = form.NewFormClient("http://localhost:1080")
	c.mock = msc.NewMockServer(msc.Config{
		Host:    "localhost",
		Port:    1080,
		Verbose: true,
	})
}

func (c *FormClientSuite) TestFormSubmission() {
	data := map[string][]string{}
	e := c.mock.On(http.MethodPost, "/form/submit").
		DefaultResponse(msc.WithStatusCode(http.StatusAccepted)).
		NumCalls(1).
		AssertionAtCall(0, msc.NewAssertion().
			WithFormURLEncodedBody(data).
			AddHeader("Content-Type", "application/x-www-form-urlencoded"),
		)

	c.Require().NoError(c.mock.Setup(context.Background(), e))

	c.Require().NoError(
		c.client.Submit(map[string]string{
			"username": "John",
			"password": "Doe",
		}),
	)

	c.Require().NoError(c.mock.Verify(context.Background(), c.T()))

	actualUserName, ok := data["username"]
	c.Require().True(ok)
	c.Require().Len(actualUserName, 1)
	c.Equal("John", actualUserName[0])
}

func (c *FormClientSuite) SetupTest() {
	c.Require().NoError(c.mock.Reset(context.Background()))
}

func TestFormClientSuite(t *testing.T) {
	suite.Run(t, &FormClientSuite{})
}
