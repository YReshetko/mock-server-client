package mock_server_client

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type bodyDecoder func(interface{}) error

type assertion struct {
	bodyDecoder          bodyDecoder
	requireBodyAssertion bool

	headers       map[string]string
	headersRegexp map[string]*regexp.Regexp
	noHeaders     map[string]struct{}

	queryParams      map[string]string
	queryParamRegexp map[string]*regexp.Regexp
	noQueryParams    map[string]struct{}

	path                 string
	requirePathAssertion bool
}

// NewAssertion creates new assertion to be checked on MockServer.Verify...
func NewAssertion() *assertion {
	return &assertion{}
}

// WithJsonBody setup HTTP request JSON body unmarshaler. The expectedBody will be prefilled on MockServer.Verify...
// and can be checked during testing.
// For example:
// 		b := SomeBody{}
// 		expectation := serverMock.On(http.SomeMethod, "/some/endpoint").
//   		AssertionAtCall(0, msc.NewAssertion().WithJsonBody(&b)
// 		someService.Process(...)
// 		...
// 		serverMock.Verify(ctx, t)
// 		assert.Equal(t, "some-value", b.SomeField)
func (a *assertion) WithJsonBody(expectedBody interface{}) *assertion {
	a.bodyDecoder = func(actualBody interface{}) error {
		switch body := actualBody.(type) {
		case []byte:
			if !json.Valid(body) {
				return fmt.Errorf("request %s was invalid json", string(body))
			}
			err := json.Unmarshal(body, expectedBody)
			if err != nil {
				return err
			}
		case string:
			if !json.Valid([]byte(body)) {
				return fmt.Errorf("request %s was invalid json", body)
			}
			err := json.Unmarshal([]byte(body), expectedBody)
			if err != nil {
				return err
			}
		default:
			m := map[string]interface{}{}
			bytes, err := json.Marshal(actualBody)
			if err != nil {
				return fmt.Errorf("unable to marshal request %v: %w", actualBody, err)
			}
			err = json.Unmarshal(bytes, &m)
			if err != nil {
				return fmt.Errorf("unable to unmarshal request %v: %w", actualBody, err)
			}

			switch m["type"] {
			case "JSON":
				detectedBody, ok := m["json"]
				if !ok {
					return fmt.Errorf("no string parameters in json [%v]", actualBody)
				}
				bytes, err := json.Marshal(detectedBody)
				if err != nil {
					return fmt.Errorf("unable to marshal request %v: %w", detectedBody, err)
				}
				err = json.Unmarshal(bytes, expectedBody)
				if err != nil {
					return fmt.Errorf("unable to unmarshal request %v: %w", detectedBody, err)
				}

			default:
				return fmt.Errorf("unknown json type, expected one of [%v]; got %v", []string{"JSON"}, m["type"])
			}
			return fmt.Errorf("request %v is not in expected format", body)
		}
		return nil
	}
	a.requireBodyAssertion = true
	return a
}

// WithFormURLEncodedBody setup HTTP request form body unmarshaler. The expectedBody will be prefilled on MockServer.Verify...
// and can be checked during testing.
func (a *assertion) WithFormURLEncodedBody(expectedBody map[string][]string) *assertion {
	a.bodyDecoder = func(actualBody interface{}) error {
		m := map[string]interface{}{}
		bytes, err := json.Marshal(actualBody)
		if err != nil {
			return fmt.Errorf("unable to marshal request %v: %w", actualBody, err)
		}
		err = json.Unmarshal(bytes, &m)
		if err != nil {
			return fmt.Errorf("unable to unmarshal request %v: %w", actualBody, err)
		}

		switch m["type"] {
		case "STRING":
			line, ok := m["string"]
			if !ok {
				return fmt.Errorf("no string parameters in form [%v]", actualBody)
			}
			pairs := strings.Split(line.(string), "&")
			for _, pair := range pairs {
				kv := strings.Split(pair, "=")
				if len(kv) != 2 {
					return fmt.Errorf("form parameter violates requirements %v", kv)
				}
				params, ok := expectedBody[kv[0]]
				if !ok {
					params = []string{}
				}
				params = append(params, kv[1])
				expectedBody[kv[0]] = params
			}

		default:
			return fmt.Errorf("unknown form type, expected one of [%v]; got %v", []string{"STRING"}, m["type"])
		}
		return nil
	}
	a.requireBodyAssertion = true
	return a
}

// WithPlainTextBody setup HTTP request plain text body unmarshaler. The expectedBody will be prefilled on MockServer.Verify...
// and can be checked during testing.
func (a *assertion) WithPlainTextBody(expectedBody *string) *assertion {
	a.bodyDecoder = func(actualBody interface{}) error {
		switch body := actualBody.(type) {
		case []byte:
			*expectedBody = string(body)
		case string:
			*expectedBody = body
		default:
			return fmt.Errorf("expected body can not be interpretade as string %v", actualBody)
		}
		return nil
	}
	a.requireBodyAssertion = true
	return a
}

// WithNoBody setup HTTP requests with no body. The expectation will fail on MockServer.Verify...
// if system sends request with body.
func (a *assertion) WithNoBody() *assertion {
	a.bodyDecoder = func(actualBody interface{}) error {
		if actualBody != nil {
			return fmt.Errorf("expected no body, but got %v", actualBody)
		}
		return nil
	}
	a.requireBodyAssertion = true
	return a
}

// WithHeaders sets required headers which sent within HTTP request to mock server.
// Sets exactly those headers which were passed to the method.
// Verifies exact match of header value.
func (a *assertion) WithHeaders(h map[string]string) *assertion {
	a.headers = h
	return a
}

// AddHeaders adds required headers which sent within HTTP request to mock server.
// Verifies exact match of header value.
func (a *assertion) AddHeaders(h map[string]string) *assertion {
	if a.headers == nil {
		return a.WithHeaders(h)
	}
	for k, v := range h {
		a.headers[k] = v
	}
	return a
}

// AddHeader adds a single header which sent within HTTP request to mock server.
// Verifies exact match of header value.
func (a *assertion) AddHeader(key, value string) *assertion {
	if a.headers == nil {
		return a.WithHeaders(map[string]string{key: value})
	}
	a.headers[key] = value
	return a
}

// AddHeaderRegexp adds a single header which sent within HTTP request to mock server.
// Verifies header by regexp.
func (a *assertion) AddHeaderRegexp(key, value string) *assertion {
	if a.headersRegexp == nil {
		a.headersRegexp = map[string]*regexp.Regexp{}
	}
	a.headersRegexp[key] = regexp.MustCompile(value)
	return a
}

// WithNoHeader verifies that the particular header was not sent to mock server app withing particular request.
func (a *assertion) WithNoHeader(key string) *assertion {
	if a.noHeaders == nil {
		a.noHeaders = map[string]struct{}{}
	}
	a.noHeaders[key] = struct{}{}
	return a
}

// WithQueryParameters sets required query parameters which sent within HTTP request to mock server.
// Sets exactly those query parameters that have to be validated.
// Verifies exact match of query parameters.
func (a *assertion) WithQueryParameters(p map[string]string) *assertion {
	a.queryParams = p
	return a
}

// AddQueryParameters adds required query parameters which sent within HTTP request to mock server.
// Verifies exact match of query parameters.
func (a *assertion) AddQueryParameters(p map[string]string) *assertion {
	if a.queryParams == nil {
		return a.WithQueryParameters(p)
	}
	for k, v := range p {
		a.queryParams[k] = v
	}
	return a
}

// AddQueryParameter adds a single query parameter which sent within HTTP request to mock server.
// Verifies exact match of query parameters.
func (a *assertion) AddQueryParameter(key, value string) *assertion {
	if a.queryParams == nil {
		return a.WithQueryParameters(map[string]string{key: value})
	}
	a.queryParams[key] = value
	return a
}

// AddQueryParamRegexp adds a single query parameter which sent within HTTP request to mock server.
// Verifies query parameter by regexp.
func (a *assertion) AddQueryParamRegexp(key, value string) *assertion {
	if a.queryParamRegexp == nil {
		a.queryParamRegexp = map[string]*regexp.Regexp{}
	}
	a.queryParamRegexp[key] = regexp.MustCompile(value)
	return a
}

// WithNoQueryParameter verifies that the particular query parameter was not sent to mock server app withing particular request.
func (a *assertion) WithNoQueryParameter(key string) *assertion {
	if a.noQueryParams == nil {
		a.noQueryParams = map[string]struct{}{}
	}
	a.noQueryParams[key] = struct{}{}
	return a
}

// WithPath verifies exact endpoint path which was called on mock server app withing particular request.
func (a *assertion) WithPath(path string) *assertion {
	a.path = path
	a.requirePathAssertion = true
	return a
}

type verification struct {
	path        string
	queryParams map[string][]string
	headers     map[string]interface{}
	body        interface{}
}

func (v *verification) assertBody(decoder bodyDecoder) error {
	return decoder(v.body)
}

func (v *verification) assertHeader(key string, value string) error {
	actualValue, ok := v.headers[key]
	if !ok {
		return fmt.Errorf("no expected header: %s", key)
	}
	switch av := actualValue.(type) {
	case []interface{}:
		for _, s := range av {
			if s == value {
				return nil
			}
		}
		return fmt.Errorf("for header %s expected value %s; actual values %v", key, value, av)
	case interface{}:
		if av == value {
			return nil
		}
		return fmt.Errorf("for header %s expected value %s; actual value %s", key, value, av)
	default:
		return fmt.Errorf("expected header %s is not a string %+v", key, av)
	}
}

func (v *verification) assertHeaderRegexp(key string, r *regexp.Regexp) error {
	actualValue, ok := v.headers[key]
	if !ok {
		return fmt.Errorf("no expected query parameter: %s", key)
	}
	switch av := actualValue.(type) {
	case string:
		if r.MatchString(av) {
			return nil
		}
		return fmt.Errorf("header parameter %s : %s does not mathc pattern: %s", key, av, r.String())
	case []string:
		for _, s := range av {
			if r.MatchString(s) {
				return nil
			}
		}
		return fmt.Errorf("for header %s : %v no one mathces pattern: %s", key, actualValue, r.String())
	case []interface{}:
		for _, h := range av {
			header, ok := h.(string)
			if !ok {
				continue
			}
			if r.MatchString(header) {
				return nil
			}
		}
		return fmt.Errorf("for header %s : %v no one mathces pattern: %s", key, actualValue, r.String())
	case interface{}:
		header, ok := av.(string)
		if ok && r.MatchString(header) {
			return nil
		}
		return fmt.Errorf("for header %s : %v no one mathces pattern: %s", key, actualValue, r.String())
	default:
		return fmt.Errorf("unable validate header %s by regexp %s as header values are not string %v", key, r.String(), actualValue)
	}
}

func (v *verification) assertNoHeader(key string) error {
	val, ok := v.headers[key]
	if ok {
		return fmt.Errorf("unexpected header found %v for key '%s'", val, key)
	}
	return nil
}

func (v *verification) assertQueryParameter(key string, value string) error {
	actualValue, ok := v.queryParams[key]
	if !ok {
		return fmt.Errorf("no expected query parameter: %s", key)
	}

	for _, s := range actualValue {
		if s == value {
			return nil
		}
	}
	return fmt.Errorf("for query parameter %s expected value %s; actual values %v", key, value, actualValue)
}

func (v *verification) assertQueryParameterRegexp(key string, r *regexp.Regexp) error {
	actualValue, ok := v.queryParams[key]
	if !ok {
		return fmt.Errorf("no expected query parameter: %s", key)
	}

	for _, s := range actualValue {
		if r.MatchString(s) {
			return nil
		}
	}
	return fmt.Errorf("for query parameter %s : %v no one mathces pattern: %s", key, actualValue, r.String())
}

func (v *verification) assertNoQueryParameter(key string) error {
	val, ok := v.queryParams[key]
	if ok {
		return fmt.Errorf("unexpected query parameter found %v for key '%s'", val, key)
	}
	return nil
}

func (v *verification) assertPath(path string) error {
	if v.path == path {
		return nil
	}
	return fmt.Errorf("expected path %s; actual path %s", path, v.path)
}
