package mock_server_client

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/YReshetko/mock-server-client/internal/client"
)

var _ MockServer = (*mockServer)(nil)

// MockServer to communicate with mock server application https://www.mock-server.com/#what-is-mockserver.
// It contains methods to create Expectation, Setup that on server and Verify any assertion during testing.
// When the particular test completed Reset method should be called to avoid other tests invalid verification results.
// During test execution somme expectation can be removed by Clear method.
type MockServer interface {
	On(method, path string) *Expectation

	Setup(context.Context, ...*Expectation) error

	Verify(context.Context, *testing.T) error
	VerifyExpectation(context.Context, *testing.T, *Expectation) error

	Clear(context.Context, *Expectation) error
	Reset(context.Context) error
}

// Config to communicate with mock server app.
type Config struct {
	Host    string
	Port    int
	Verbose bool
}

type mockServer struct {
	client client.Client

	expectations map[string]*Expectation
}

// NewMockServer creates a new MockServer client
func NewMockServer(cfg Config) *mockServer {
	return &mockServer{
		client:       client.NewClient(cfg.Host, cfg.Port, cfg.Verbose),
		expectations: map[string]*Expectation{},
	}
}

// On creates new Expectation when testing requires call to external endpoint by some HTTP method.
// Expectation itself is builder, so you can set up it accordingly using corresponding approach:
// expectation.Name("someName").NumCalls(10).Request(...)...
func (m *mockServer) On(method, path string) *Expectation {
	e := newExpectation(method, path)
	m.expectations[e.id] = &e
	return &e
}

// Setup initialises Expectation on mock server app.
// It sends expectation request to corresponding mock server app.
func (m *mockServer) Setup(ctx context.Context, expectations ...*Expectation) error {
	for _, expectation := range expectations {
		for _, e := range expectation.build() {
			err := m.client.Expectation(ctx, e)
			if err != nil {
				// TODO make debuggable invalid expectations
				return errors.Wrapf(err, "unable to setup expectation %s", e.ID)
			}
		}
	}
	return nil
}

// Verify checks all []Expectation which were created by On method.
func (m *mockServer) Verify(ctx context.Context, t *testing.T) error {
	for _, expectation := range m.expectations {
		err := m.verifyExpectation(ctx, t, expectation)
		if err != nil {
			return errors.Wrapf(err, "verification failed on expectation %s", expectation.id)
		}
	}
	return nil
}

// VerifyExpectation checks the Expectation which was sent directly to the method.
func (m *mockServer) VerifyExpectation(ctx context.Context, t *testing.T, expectation *Expectation) error {
	return m.verifyExpectation(ctx, t, expectation)
}

func (m *mockServer) verifyExpectation(ctx context.Context, t *testing.T, expectation *Expectation) error {
	if expectation.numCalls == 0 && len(expectation.assertions) == 0 {
		return nil
	}
	name := expectation.String()

	verifications, err := m.verifications(ctx, expectation)
	if err != nil {
		t.Fatalf("FAIL assertion:\n"+
			"Expectation name [%s]\n"+
			"Reason: unable to verify expectation %s", name, expectation.id)
	}

	if expectation.numCalls != 0 {
		if len(verifications) != expectation.numCalls {
			t.Fatalf("FAIL assertion:\n"+
				"Expectation name [%s]\n"+
				"Reason: expected num calls to %s: %d; actual: %d", name, expectation.request.path, expectation.numCalls, len(verifications))
		}
	}

	fail := false

	asserErr := func(id int, err error) {
		if err != nil {
			t.Errorf("FAIL assertion:\n"+
				"Expectation name [%s]\n"+
				"Assertion at call [%d]\n"+
				"Reason: %s", name, id, err)
			fail = true
		}
	}

	for i, a := range expectation.assertions {
		if i >= len(verifications) {
			asserErr(i, fmt.Errorf("assertion index %d is out of bounds made calls %d", i, len(verifications)))
			fail = true
			continue
		}
		ver := verifications[i]

		if a.requireBodyAssertion {
			asserErr(i, ver.assertBody(a.bodyDecoder))
		}

		if a.requirePathAssertion {
			asserErr(i, ver.assertPath(a.path))
		}

		for k, v := range a.queryParamRegexp {
			asserErr(i, ver.assertQueryParameterRegexp(k, v))
		}
		for k, v := range a.queryParams {
			asserErr(i, ver.assertQueryParameter(k, v))
		}
		for k, _ := range a.noQueryParams {
			asserErr(i, ver.assertNoQueryParameter(k))
		}

		for k, v := range a.headers {
			asserErr(i, ver.assertHeader(k, v))
		}
		for k, v := range a.headersRegexp {
			asserErr(i, ver.assertHeaderRegexp(k, v))
		}
		for k, _ := range a.noHeaders {
			asserErr(i, ver.assertNoHeader(k))
		}

	}

	if fail {
		t.FailNow()
	}

	return nil
}

func (m *mockServer) verifications(ctx context.Context, expectation *Expectation) ([]verification, error) {
	rq := clientHttpRequest(expectation.request)
	rs, err := m.client.Retrieve(ctx, client.RetrieveRequest(rq))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get verifications for expectation %s", expectation.id)
	}

	v := make([]verification, len(rs))
	for i, r := range rs {
		v[i] = verification{
			path:        r.Path,
			queryParams: r.QueryStringParameters,
			headers:     r.Headers,
			body:        r.Body,
		}
	}
	return v, nil
}

// Clear removes the Expectation from mock server app and unregister it on MockServer client.
func (m *mockServer) Clear(ctx context.Context, expectation *Expectation) error {
	err := m.client.Clear(ctx, client.ClearRequest{ExpectationID: client.ExpectationID{ID: expectation.id}})
	if err != nil {
		return errors.Wrapf(err, "unable to remove expectation %s", expectation.id)
	}
	delete(m.expectations, expectation.id)
	return nil
}

// Reset removes all []Expectation from mock server app and MockServer client.
func (m *mockServer) Reset(ctx context.Context) error {
	err := m.client.Reset(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to reset mock server expectations")
	}
	m.expectations = map[string]*Expectation{}
	return nil
}
