package mock_server_client

import "time"

type RequestOption func(*request)

// WithPathParameter sets required request path parameter that has to be checked on mock
// server app to return corresponding response.
func WithPathParameter(key, value string) RequestOption {
	return func(r *request) {
		if r.pathParams == nil {
			r.pathParams = map[string]string{}
		}
		r.pathParams[key] = value
	}
}

// WithQueryParameter sets required query parameter that has to be checked on mock
// server app to return corresponding response.
func WithQueryParameter(key, value string) RequestOption {
	return func(r *request) {
		if r.queryParams == nil {
			r.queryParams = map[string]string{}
		}
		r.queryParams[key] = value
	}
}

// WithRequestHeader sets required header parameter that has to be checked on mock
// server app to return corresponding response.
func WithRequestHeader(key, value string) RequestOption {
	return func(r *request) {
		if r.headers == nil {
			r.headers = map[string]string{}
		}
		r.headers[key] = value
	}
}

// WithRequestBody sets required body that has to be checked on mock
// server app to return corresponding response.
func WithRequestBody(body interface{}) RequestOption {
	return func(r *request) {
		r.body = body
	}
}

type ResponseOption func(*response)

// WithResponseHeader sets response header to be returned within corresponding HTTP response from mock server app.
func WithResponseHeader(key, value string) ResponseOption {
	return func(r *response) {
		if r.headers == nil {
			r.headers = map[string]string{}
		}
		r.headers[key] = value
	}
}

// WithResponseBody sets response body to be returned within corresponding HTTP response from mock server app.
func WithResponseBody(body interface{}) ResponseOption {
	return func(r *response) {
		r.body = body
	}
}

// WithStatusCode sets HTTP status code to be returned within corresponding HTTP response from mock server app.
func WithStatusCode(s int) ResponseOption {
	return func(r *response) {
		r.statusCode = s
	}
}

// WithDelay sets HTTP response delay for mock server app.
func WithDelay(s time.Duration) ResponseOption {
	return func(r *response) {
		r.delay = &s
	}
}

// WithReason sets HTTP failure reason for mock server app.
func WithReason(s string) ResponseOption {
	return func(r *response) {
		r.reasonPhrase = s
	}
}

// WithDropConnection sets true to brake HTTP connection from mock server app side for particular response.
func WithDropConnection() ResponseOption {
	return func(r *response) {
		r.drop = true
	}
}

// WithErrorBytes sets error bytes within dropped connection on mock server side.
func WithErrorBytes(b []byte) ResponseOption {
	return func(r *response) {
		r.errorBytes = b
	}
}
