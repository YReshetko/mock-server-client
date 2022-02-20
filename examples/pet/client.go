package pet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const (
	allPetsPath = "/pets"
	petByIDPath = "/pets/%d"
)

type Pet struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Pets []Pet

var NotFoundError = errors.New("pet not found")

type PetClient struct {
	client *http.Client
	host   string
}

func NewPetClient(host string) *PetClient {
	return &PetClient{
		client: http.DefaultClient,
		host:   host,
	}
}

func (c *PetClient) All() (Pets, error) {
	rq, err := http.NewRequest(http.MethodGet, c.host+allPetsPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get all pets")
	}
	rs, err := c.client.Do(rq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to call API to get all pets")
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read pets body")
	}

	var pets Pets
	err = json.Unmarshal(body, &pets)
	return pets, errors.Wrap(err, "unable to unmarshal pets response")
}

func (c *PetClient) ByID(id int) (Pet, error) {
	rq, err := http.NewRequest(http.MethodGet, fmt.Sprintf(c.host+petByIDPath, id), nil)
	if err != nil {
		return Pet{}, errors.Wrap(err, "unable to get all pets")
	}
	rs, err := c.client.Do(rq)
	if err != nil {
		return Pet{}, errors.Wrap(err, "unable to call API to get pet by ID")
	}
	defer rs.Body.Close()
	if rs.StatusCode == http.StatusNotFound {
		return Pet{}, NotFoundError
	}

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return Pet{}, errors.Wrap(err, "unable to read pets body")
	}

	var pet Pet
	err = json.Unmarshal(body, &pet)
	return pet, errors.Wrap(err, "unable to unmarshal pets response")
}

func (c *PetClient) Add(pet Pet) error {
	data, err := json.Marshal(pet)
	if err != nil {
		return errors.Wrap(err, "unable to marshal pet")
	}
	rq, err := http.NewRequest(http.MethodPost, c.host+allPetsPath, bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "unable to get all pets")
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		return errors.Wrap(err, "unable to call API to add pet")
	}
	defer rs.Body.Close()

	if rs.StatusCode < 200 || rs.StatusCode > 299 {
		return errors.New(fmt.Sprintf("expected 2xx, got %d", rs.StatusCode))
	}
	return nil
}
