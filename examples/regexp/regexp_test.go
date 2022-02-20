package regexp_test

import (
	"context"
	"github.com/YReshetko/mock-server-client/examples/regexp"
	"github.com/google/uuid"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	msc "github.com/YReshetko/mock-server-client"
)

type RegexpClientSuite struct {
	suite.Suite

	client *regexp.Client
	mock   msc.MockServer
}

func (c *RegexpClientSuite) SetupSuite() {
	c.client = regexp.NewClient("http://localhost:1080")
	c.mock = msc.NewMockServer(msc.Config{
		Host:    "localhost",
		Port:    1080,
		Verbose: true,
	})
}

func (c *RegexpClientSuite) TestHeaderRegexps() {
	e := c.mock.On(http.MethodGet, "/some/endpoint").
		Name("Submit headers").
		DefaultResponse(msc.WithStatusCode(http.StatusAccepted)).
		NumCalls(1).
		AssertionAtCall(0, msc.NewAssertion().
			WithNoBody().
			AddQueryParamRegexp("option", "^dev_.+_\\d{1,2}$").
			//WithNoQueryParameter("option").
			WithNoQueryParameter("foo").
			AddHeaderRegexp("X-User-Id", "[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}").
			AddHeaderRegexp("X-User-Role", ".+_.+_ADMIN").
			AddHeaderRegexp("Authorization", "Bearer .+\\..+\\..+").
			//WithNoHeader("X-User-Id").
			WithNoHeader("X-Unexpected-Header"),
		)

	c.Require().NoError(c.mock.Setup(context.Background(), e))

	uuid.NewString()
	c.Require().NoError(
		c.client.Do(map[string][]string{
			"X-User-Id":     {uuid.NewString()},
			"X-User-Role":   {"READ_DATA_ADMIN", "READ_DATA_USER"},
			"Authorization": {"Bearer n347yr3cy8f7c4y3b38y9384.c3n874yr9834ybr8cb8y398jdr834.rnc3874y8r943ybcby8"},
		}, "option=dev_modifier_14&option=admin_change_67"),
	)

	c.Require().NoError(c.mock.VerifyExpectation(context.Background(), c.T(), e))
}

func (c *RegexpClientSuite) TestAuth() {
	e := c.mock.On(http.MethodPost, "/some/endpoint").
		DefaultResponse(msc.WithStatusCode(http.StatusOK)).
		NumCalls(1).
		AssertionAtCall(0, msc.NewAssertion().
			WithNoBody().
			AddHeaderRegexp("Authorization", "Basic .+"),
		//AddHeaderRegexp("Authorization", "Bearer .+"),
		)

	c.Require().NoError(c.mock.Setup(context.Background(), e))

	uuid.NewString()
	c.Require().NoError(
		c.client.BasicAuth("john", "doe"),
	)

	c.Require().NoError(c.mock.Verify(context.Background(), c.T()))
}

func (c *RegexpClientSuite) SetupTest() {
	c.Require().NoError(c.mock.Reset(context.Background()))
}

func TestRegexpClientSuite(t *testing.T) {
	suite.Run(t, &RegexpClientSuite{})
}
