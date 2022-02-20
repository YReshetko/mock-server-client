package mock_server_client

import (
	"time"

	"github.com/google/uuid"

	"github.com/YReshetko/mock-server-client/internal/client"
)

// Expectation contains references to expected HTTP request and responses, required assertions and so on.
// Each point can not be checked directly, but using AssertionAtCall and setup proper assertions.
// The verification can be done at any point by MockServer.Verify or MockServer.VerifyExpectation during testing.
type Expectation struct {
	id                  string
	name                string
	request             *request
	defaultResponse     *response
	sequentialResponses []response
	assertions          map[int]*assertion
	numCalls            int

	isBuilt bool
}

type request struct {
	method      string
	path        string
	pathParams  map[string]string
	queryParams map[string]string
	headers     map[string]string
	body        interface{}
}

type response struct {
	body         interface{}
	statusCode   int
	reasonPhrase string
	headers      map[string]string
	delay        *time.Duration
	drop         bool
	errorBytes   []byte
}

func newExpectation(method, path string) Expectation {
	return Expectation{
		id: uuid.NewString(),
		request: &request{
			method: method,
			path:   path,
		},
	}
}

// Request prepares expected request.
// Can not be called after the Expectation was MockServer.Setup to mock server app, it leads the panic().
func (e *Expectation) Request(opts ...RequestOption) *Expectation {
	if e.isBuilt {
		panic("unable to update assertion when it's already on mock server")
	}
	for _, opt := range opts {
		opt(e.request)
	}
	return e
}

// DefaultResponse prepares default response which is returned by mock server up when there is no SequentialResponse,
// or calls to all SequentialResponse are completed.
// Can not be called after the Expectation was MockServer.Setup to mock server app, it leads the panic().
func (e *Expectation) DefaultResponse(opts ...ResponseOption) *Expectation {
	if e.isBuilt {
		panic("unable to update assertion when it's already on mock server")
	}
	r := &response{}
	for _, opt := range opts {
		opt(r)
	}
	e.defaultResponse = r
	return e
}

// SequentialResponse prepares ordered responses which can be returned only once in the same order as
// SequentialResponse was called on Expectation.
// Can not be called after the Expectation was MockServer.Setup to mock server app, it leads the panic().
func (e *Expectation) SequentialResponse(opts ...ResponseOption) *Expectation {
	if e.isBuilt {
		panic("unable to update assertion when it's already on mock server")
	}
	r := response{}
	for _, opt := range opts {
		opt(&r)
	}
	e.sequentialResponses = append(e.sequentialResponses, r)
	return e
}

// NumCalls specific assertion to check if corresponding request was sent exactly N times.
func (e *Expectation) NumCalls(value int) *Expectation {
	e.numCalls = value
	return e
}

// AssertionAtCall register the number of assertions which has to be done on the Expectation for particular callNumber.
// NewAssertion() creates an assertion that is designed with builder pattern, so it can be prepared accordingly:
// assertion.WithNoBody().AddHeader("key", "value")...
func (e *Expectation) AssertionAtCall(callNumber int, a *assertion) *Expectation {
	if e.assertions == nil {
		e.assertions = map[int]*assertion{}
	}
	e.assertions[callNumber] = a
	return e
}

// Name set Expectation name for better debug, if the name is not set the new UUID will be generated instead.
// For example:
// 		Test failure on named Expectation:
//			FAIL assertion:
//        		Expectation name [Some user freandly name]
//        		Assertion at call [0]
//        		Reason: unexpected query parameter found [dev_modifier_14 admin_change_67] for key 'option'
//
// 		Test failure on random Expectation:
// 			FAIL assertion:
//        		Expectation name [1cf2db7f-51c4-4961-be1a-de361a7c0db8]
//        		Assertion at call [0]
//        		Reason: unexpected query parameter found [dev_modifier_14 admin_change_67] for key 'option'
func (e *Expectation) Name(name string) *Expectation {
	e.name = name
	return e
}

// String returns Expectation name or id if the name was not added.
func (e *Expectation) String() string {
	if e.name != "" {
		return e.name
	}
	return e.id
}

func (e *Expectation) build() []client.Expectation {
	if e.isBuilt {
		panic("unable to build assertion more then once")
	}
	e.isBuilt = true
	expectations := make([]client.Expectation, len(e.sequentialResponses)+1)
	httpRequest := clientHttpRequest(e.request)

	for i, response := range e.sequentialResponses {
		exp := newClientExpectation(&response)
		exp.Times = &client.Times{
			RemainingTimes: 1,
			Unlimited:      false,
		}
		exp.Priority = len(expectations) - i
		exp.HTTPRequest = &httpRequest
		expectations[i] = exp
	}

	defaultExp := newClientExpectation(e.defaultResponse)
	defaultExp.Times = &client.Times{
		Unlimited: true,
	}
	defaultExp.HTTPRequest = &httpRequest
	expectations[len(expectations)-1] = defaultExp
	return expectations
}

func newClientExpectation(res *response) client.Expectation {
	e := client.Expectation{
		ID: uuid.NewString(),
		TimeToLive: &client.TimeToLive{
			Unlimited: true,
		},
	}

	if res.drop {
		e.HTTPError = &client.HTTPError{
			Delay:          delay(res.delay),
			DropConnection: true,
			ResponseBytes:  string(res.errorBytes),
		}
	} else {
		e.HTTPResponse = &client.HTTPResponse{
			Body:         res.body,
			StatusCode:   res.statusCode,
			ReasonPhrase: res.reasonPhrase,
			Headers:      toClientHeaders(res.headers),
			Delay:        delay(res.delay),
		}
	}

	return e
}

func delay(t *time.Duration) *client.Delay {
	if t == nil || *t == 0 {
		return nil
	}
	return &client.Delay{
		TimeUnit: client.MILLISECONDS,
		Value:    int(t.Milliseconds()),
	}
}

func clientHttpRequest(req *request) client.HTTPRequest {
	return client.HTTPRequest{
		Method:                req.method,
		Path:                  req.path,
		PathParameters:        toClientMap(req.pathParams),
		QueryStringParameters: toClientMap(req.queryParams),
		Headers:               toClientHeaders(req.headers),
		Body:                  req.body,
	}
}

func toClientMap(m map[string]string) map[string][]string {
	out := map[string][]string{}
	for k, v := range m {
		out[k] = []string{v}
	}
	return out
}

func toClientHeaders(m map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range m {
		out[k] = []string{v}
	}
	return out
}
