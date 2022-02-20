package client

// Expectation

type Expectation struct {
	ID           string        `json:"id"`
	Priority     int           `json:"priority"`
	HTTPRequest  *HTTPRequest  `json:"httpRequest,omitempty"`
	HTTPResponse *HTTPResponse `json:"httpResponse,omitempty"`
	Times        *Times        `json:"times,omitempty"`
	TimeToLive   *TimeToLive   `json:"timeToLive,omitempty"`
	HTTPError    *HTTPError    `json:"httpError,omitempty"`
}

type HTTPRequest struct {
	Method                string                 `json:"method"`
	Path                  string                 `json:"path"`
	PathParameters        map[string][]string    `json:"pathParameters,omitempty"`
	QueryStringParameters map[string][]string    `json:"queryStringParameters,omitempty"`
	Headers               map[string]interface{} `json:"headers,omitempty"`
	Body                  interface{}            `json:"body,omitempty"`
}

type HTTPResponse struct {
	Body         interface{}            `json:"body,omitempty"`
	StatusCode   int                    `json:"statusCode"`
	ReasonPhrase string                 `json:"reasonPhrase,omitempty"`
	Headers      map[string]interface{} `json:"headers,omitempty"`
	Delay        *Delay                 `json:"delay,omitempty"`
}

type HTTPError struct {
	Delay          *Delay `json:"delay,omitempty"`
	DropConnection bool   `json:"dropConnection"`
	ResponseBytes  string `json:"responseBytes,omitempty"`
}

type Times struct {
	RemainingTimes int  `json:"remainingTimes"`
	Unlimited      bool `json:"unlimited"`
}

type TimeUnit string

const (
	DAYS         TimeUnit = "DAYS"
	HOURS        TimeUnit = "HOURS"
	MINUTES      TimeUnit = "MINUTES"
	SECONDS      TimeUnit = "SECONDS"
	MILLISECONDS TimeUnit = "MILLISECONDS"
	MICROSECONDS TimeUnit = "MICROSECONDS"
	NANOSECONDS  TimeUnit = "NANOSECONDS"
)

type TimeToLive struct {
	TimeUnit   TimeUnit `json:"timeUnit,omitempty"`
	TimeToLive int      `json:"timeToLive,omitempty"`
	Unlimited  bool     `json:"unlimited"`
}

type Delay struct {
	TimeUnit TimeUnit `json:"timeUnit"`
	Value    int      `json:"value"`
}

// Verify

type Verify struct {
	ExpectationID ExpectationID      `json:"expectationId"`
	HTTPRequest   *HTTPRequest       `json:"httpRequest,omitempty"`
	Times         *VerificationTimes `json:"times,omitempty"`
}

type ExpectationID struct {
	ID string `json:"id"`
}

type VerificationTimes struct {
	AtLeast int `json:"atLeast"`
	AtMost  int `json:"atMost"`
}

type VerifySequence struct {
	ExpectationIDs []ExpectationID `json:"expectationIds,omitempty"`
	HTTPRequests   *HTTPRequest    `json:"httpRequest,omitempty"`
}

type RetrieveRequest HTTPRequest

type RetrieveResponse []HTTPRequest

// Clear

type ClearRequest struct {
	ExpectationID ExpectationID `json:"expectationId"`
}
