package pet_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	msc "github.com/YReshetko/mock-server-client"
	"github.com/YReshetko/mock-server-client/examples/pet"
)

type PetClientSuite struct {
	suite.Suite

	client *pet.PetClient
	mock   msc.MockServer
}

func (c *PetClientSuite) SetupSuite() {
	c.client = pet.NewPetClient("http://localhost:1080")
	c.mock = msc.NewMockServer(msc.Config{
		Host:    "localhost",
		Port:    1080,
		Verbose: true,
	})
}

func (c *PetClientSuite) SetupTest() {
	c.Require().NoError(c.mock.Reset(context.Background()))
}

func (c *PetClientSuite) TestGetAll() {
	expectation := c.mock.On(http.MethodGet, "/pets").DefaultResponse(
		msc.WithStatusCode(http.StatusOK),
		msc.WithResponseBody(pet.Pets{{Name: "JoJo", Age: 2}, {Name: "PoPo", Age: 3}}),
	)
	c.Require().NoError(c.mock.Setup(context.Background(), expectation))

	pets, err := c.client.All()
	c.Require().NoError(err)
	c.Len(pets, 2)

	c.Equal("JoJo", pets[0].Name)
	c.Equal(2, pets[0].Age)

	c.Equal("PoPo", pets[1].Name)
	c.Equal(3, pets[1].Age)
}

func (c *PetClientSuite) TestSequentialRead() {
	expectation := c.mock.On(http.MethodGet, "/pets/{pet_id}").
		Request(
			msc.WithPathParameter("pet_id", "[0-9]{1}"),
		).
		SequentialResponse(
			msc.WithStatusCode(http.StatusOK),
			msc.WithResponseBody(pet.Pet{Name: "JoJo", Age: 1}),
		).
		SequentialResponse(
			msc.WithStatusCode(http.StatusOK),
			msc.WithResponseBody(pet.Pet{Name: "PoPo", Age: 4}),
		).
		DefaultResponse(
			msc.WithStatusCode(http.StatusNotFound),
		).
		//NumCalls(5).
		NumCalls(4).
		AssertionAtCall(0, msc.NewAssertion().
			WithNoBody().
			WithPath("/pets/1"),
		)
	c.Require().NoError(c.mock.Setup(context.Background(), expectation))

	p, err := c.client.ByID(1)
	c.Require().NoError(err)
	c.Equal("JoJo", p.Name)
	c.Equal(1, p.Age)

	p, err = c.client.ByID(2)
	c.Require().NoError(err)
	c.Equal("PoPo", p.Name)
	c.Equal(4, p.Age)

	p, err = c.client.ByID(3)
	c.Require().Error(err)
	c.Equal(pet.NotFoundError, err)

	p, err = c.client.ByID(4)
	c.Require().Error(err)
	c.Equal(pet.NotFoundError, err)

	err = c.mock.VerifyExpectation(context.Background(), c.T(), expectation)
	c.Require().NoError(err)
}

func (c *PetClientSuite) TestSequentialCreation() {
	expectedPetRequests := [3]pet.Pet{}

	expectation := c.mock.On(http.MethodPost, "/pets").
		SequentialResponse(
			msc.WithStatusCode(http.StatusCreated),
		).
		SequentialResponse(
			msc.WithStatusCode(http.StatusCreated),
		).
		DefaultResponse(
			msc.WithStatusCode(http.StatusInternalServerError),
		).
		NumCalls(3).
		AssertionAtCall(0, msc.NewAssertion().
			WithJsonBody(&expectedPetRequests[0]).
			//AddHeader("User-Agent", "Go-http-client/1.2").
			AddHeader("User-Agent", "Go-http-client/1.1").
			WithPath("/pets"),
		).
		AssertionAtCall(1, msc.NewAssertion().
			WithJsonBody(&expectedPetRequests[1]).
			AddHeader("User-Agent", "Go-http-client/1.1").
			WithPath("/pets"),
		).
		AssertionAtCall(2, msc.NewAssertion().
			WithJsonBody(&expectedPetRequests[2]).
			AddHeader("User-Agent", "Go-http-client/1.1").
			WithPath("/pets"),
		)
	c.Require().NoError(c.mock.Setup(context.Background(), expectation))

	c.Require().NoError(c.client.Add(pet.Pet{Name: "PoPo", Age: 5}))
	c.Require().NoError(c.client.Add(pet.Pet{Name: "JoJo", Age: 15}))
	c.Require().Error(c.client.Add(pet.Pet{Name: "LoLo", Age: 150}))

	err := c.mock.VerifyExpectation(context.Background(), c.T(), expectation)
	c.Require().NoError(err)

	c.Equal("PoPo", expectedPetRequests[0].Name)
	c.Equal("JoJo", expectedPetRequests[1].Name)
	c.Equal("LoLo", expectedPetRequests[2].Name)
}

func TestPetClientSuite(t *testing.T) {
	suite.Run(t, &PetClientSuite{})
}
