# Test client for MockServer

This is the go client with testing package integration to mock server: 

- Doc: https://www.mock-server.com/#what-is-mockserver
- Github: https://github.com/mock-server/mockserver
- Swagger: https://app.swaggerhub.com/apis/jamesdbloom/mock-server-openapi/5.12.x#/info
- Docker: https://hub.docker.com/r/mockserver/mockserver

# How to use

## Run examples

```shell
make docker-up
make examples 
```

## Write tests

1. Make sure mock server is up and running on some host (for example `localhost:1080`)
2. Initiate MockServer to start using the mock:
```go
import (
	...
    msc "github.com/YReshetko/mock-server-client"
	...
)

func TestSomething(t *testing.T) {
    mock := msc.NewMockServer(
		msc.Config{
            Host:    "localhost",
            Port:    1080,
            Verbose: true,
        },
    )
	...
}
```
3. At some point, before you expect your system calls third party service that you are going to mock, setup expectations. For example:
```go
func TestSomething(t *testing.T) {
	...
    expectRequest := SomeRequest{}
    expectation := mock.On(http.SomeMethod, "/some/{some_param}/endpoint").
        Request(
            msc.WithPathParameter("some_param", "[0-9]{1}"),
        ).
        SequentialResponse(
            msc.WithStatusCode(http.StatusCreated),
        ).
        SequentialResponse(
            msc.WithStatusCode(http.StatusOK),
        ).
        DefaultResponse(
            msc.WithStatusCode(http.StatusNotFound),
        ).
        NumCalls(3).
        AssertionAtCall(0, msc.NewAssertion().
            WithJsonBody(&expectRequest).
            AddHeader("User-Agent", "Go-http-client/1.1").
            WithPath("/some/1/endpoint"),
        )
    mock.Setup(context.Background(), expectation)
	...
}
```
4. When your system did some work, and you are going to verify sent requests to the mocked third party you need to call `Veryfy` or `VerifyExpectation` method. At this point all expected bodies that were setup in assertions will be fulfilled by the MockServer:
```go
func TestSomething(t *testing.T) {
	...
	mock.Verify(context.Background(), t)
	assert.Equal(t, "some-field-value", expectRequest.SomeField)
	...
}
```
5. The most important thing is to reset MockServer each time when you start new test scenario, otherwise all previously created expectation can affect verification result. Remember if you create a new MockServer it doesn't mean that you have cleaned the expectation on mock server app.
```go
func TestSomething(t *testing.T) {
	...
	mock.Reset(context.Background())
	...
}
```