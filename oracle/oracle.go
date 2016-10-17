package oracle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/graymeta/stow"
)

// AuthResponse encapsulates the data returned when requesting authorization
// information for accessing storage tokens.
type AuthResponse struct {
	ContentLength string `json:"Content-Length,"`
	Server        string `json:"Server,"`
	AuthToken     string `json:"X-Auth-Token,"`
	StorageToken  string `json:"X-StorageToken,"`
	StorageURL    string `json:"X-Storage-Url,"`
	Date          string `json:"date,"`
}

type Client struct {
	client   http.Client
	AuthInfo AuthResponse
}

// Get auth key
func newOracleClient(cfg stow.Config) (Client, error) {
	username, _ := cfg.Config(storageUsername)
	password, _ := cfg.Config(storagePassword)
	endpoint, _ := cfg.Config(authEndpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return Client{}, err
	}

	request.Header = make(map[string][]string)
	request.Header.Set("X-Storage-Pass", password)
	request.Header.Set("X-Storage-User", username)

	newClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := newClient.Do(request)
	if err != nil {
		return Client{}, err
	} else if response.StatusCode != 200 {
		return Client{}, fmt.Errorf("received status code %d", response.StatusCode)
	}

	var ar AuthResponse

	ar.Server = response.Header.Get("Server")
	ar.AuthToken = response.Header.Get("X-Auth-Token")
	ar.StorageURL = response.Header.Get("X-Storage-Url")

	c := Client{
		client:   *newClient,
		AuthInfo: ar,
	}

	return c, nil
}

/*
type ListContainersOutput struct {

}*/

// ListContainers sends a request to Object Storage to retrieve information about the account and containers.
// Path Parameters
// account (string, required), the unique name for the account
//
// Query Parameters
// delimeter (char string) - delimeter value, returns object names nested in the container.
// end_marker (string) - returns container names that are less than the marker value.
// format (string) - response format. Valid values are json, xml, or plain (default).
// limit (integer) - limits the number of results to n.
// marker (string) - returns container names that are greater than the marker value.
// prefix (string) - named items in the response begin with this value.
//
// Header Parameters
// Accept (string) - Instead of using the format query parameter, set this to application/json,
//                   application/xml, or text/plain.
// X-Auth-Token (string, required) - authentication token.
// X-Newest (boolen) - if true, object query replicas return the most recent result. If omitted,
//                     Object Storage response after it finds the first valid replica. Setting this
//                     explicitly to true is more expensive for the back end, so use sparingly.
//
// Response Codes
// 200  -
//
// 204 - Success, no containers. No containers, or paging through long list using marker, limit, or
//       end marker query parameter and you've reached the end of the list.
// Headers
// Content-Length (string) - Length of the response body that contains the list of names. This
//                           value is the length of the error text in the respons body on failure.
//
// 401 - Authentication tokens expire after 30 minutes
// Headers
// Content-Length (string) - The length of the error text in the response body.
// Content-Type (string) - The MIME type of the rror text in the response body.

var errAuth = errors.New("request does not include authentication token, is not valid, or may have expired")

func (c *Client) ListContainers() (*ListContainersOutput, error) {
	request, err := http.NewRequest("GET", c.AuthInfo.StorageURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("X-Auth-Token", c.AuthInfo.AuthToken)
	request.Header.Add("Accept", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	} else if response.StatusCode == 401 {
		return nil, fmt.Errorf("received status code %d, %v", response.StatusCode, errAuth)
	}

	// Parse response
	var output ListContainersOutput

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseBytes, &output.Containers)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// v1/<Account>/<container>
func (c *Client) ListItems(containerName string) (*ListContainersOutput, error) {
	endpoint := strings.Join([]string{c.AuthInfo.StorageURL, containerName}, "/")

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("X-Auth-Token", c.AuthInfo.AuthToken)
	request.Header.Add("Accept", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	} else if response.StatusCode == 401 {
		return nil, fmt.Errorf("received status code %d, %v", response.StatusCode, errAuth)
	}

	// Parse response
	//var output ListContainersOutput

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Response Bytes: %s", responseBytes)

	return nil, nil
}

