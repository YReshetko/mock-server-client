package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/YReshetko/mock-server-client/internal/client"
)

func main() {
	c := client.NewClient("localhost", 1080, true)
	//example(c)

	/*err := c.Reset(context.Background())
	if err != nil {
		log.Fatalln(err)
	}*/
	_ = returnSequentially(c)
	//verify("two", 2)

	retrieve, err := c.Retrieve(context.Background(), client.RetrieveRequest{
		Path:   "/pets",
		Method: http.MethodGet,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(retrieve)

}

func returnSequentially(c client.Client) func(id string, times int) {
	err := c.Expectation(context.Background(), client.Expectation{
		ID:       "one",
		Priority: 2,
		HTTPRequest: &client.HTTPRequest{
			Path:   "/pets",
			Method: http.MethodGet,
		},
		HTTPResponse: &client.HTTPResponse{
			StatusCode: http.StatusAccepted,
			Body: map[string]interface{}{
				"pet-1": "JoJo",
			},
		},
		Times: &client.Times{
			RemainingTimes: 1,
			Unlimited:      false,
		},
		TimeToLive: &client.TimeToLive{
			TimeUnit:   client.MINUTES,
			TimeToLive: 10,
			Unlimited:  false,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = c.Expectation(context.Background(), client.Expectation{
		ID:       "two",
		Priority: 1,
		HTTPRequest: &client.HTTPRequest{
			Path:   "/pets",
			Method: http.MethodGet,
		},
		HTTPResponse: &client.HTTPResponse{
			StatusCode: http.StatusAccepted,
			Body: map[string]interface{}{
				"pet-1": "PoPo",
			},
		},
		Times: &client.Times{
			RemainingTimes: 1,
			Unlimited:      false,
		},
		TimeToLive: &client.TimeToLive{
			TimeUnit:   client.MINUTES,
			TimeToLive: 10,
			Unlimited:  false,
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	return func(id string, times int) {
		err := c.Verify(context.Background(), client.Verify{
			ExpectationID: client.ExpectationID{
				ID: id,
			},
			Times: &client.VerificationTimes{
				AtLeast: times,
				AtMost:  times,
			},
		},
		)

		if err != nil {
			log.Fatalln(err)
		}
	}
}

func example(c client.Client) {
	err := c.Expectation(context.Background(), client.Expectation{
		ID:       "one",
		Priority: 1,
		HTTPRequest: &client.HTTPRequest{
			Path:   "/pets/{petId}",
			Method: http.MethodGet,
			PathParameters: map[string][]string{
				"petId": {"1"},
			},
			QueryStringParameters: map[string][]string{
				"hello": {"[0-9]{1}"},
			},
			Headers: map[string]interface{}{
				"keyMatchStyle":   "MATCHING_KEY",
				"Accept":          []string{"application/json"},
				"Accept-Encoding": []string{"gzip, deflate, br"},
			},
		},
		HTTPResponse: &client.HTTPResponse{
			Body: map[string]interface{}{
				"pet-1": "JoJo",
			},
		},
		Times: &client.Times{
			RemainingTimes: 1000,
			Unlimited:      true,
		},
		TimeToLive: &client.TimeToLive{
			TimeUnit:   client.MINUTES,
			TimeToLive: 10,
			Unlimited:  false,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
}
