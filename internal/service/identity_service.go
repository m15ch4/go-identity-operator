package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	v1 "github.com/m15ch4/go-identity-operator/api/v1"
)

type IdentityUser struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Password  string `json:"password,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Role      string `json:"role,omitempty"`
	Age       int    `json:"age,omitempty"`
}

type LoginRequestBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
}

type IdentityService struct {
	config *IdentityConfig
	token  string
}

func NewIdentityService(config *IdentityConfig) *IdentityService {
	return &IdentityService{
		config: config,
	}
}

// GetToken makes REST API call to /login of identity app described by config property and returns the refresh token
func (s *IdentityService) GetToken() (string, error) {
	// prepare request url
	url := "http://" + s.config.host + ":" + strconv.Itoa(s.config.port) + "/login"

	// prepare request body
	reqBody := LoginRequestBody{
		Name:     s.config.user,
		Password: s.config.pass,
	}
	// encode request body
	jsonReqBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	// prepare request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return "", err
	}

	// make rest api call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// close response body
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// extract the token field from the response body JSON object
	var loginResponse LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return "", err
	}

	// save the token in the service
	s.token = loginResponse.Token

	// return the token
	return s.token, nil
}

// CreateUser makes REST API call to /users of identity app described by config property and returns the IdentityUser object.
// Request's body contains IdentityUser in JSON format.
// REST API call uses POST HTTP method.
func (s *IdentityService) CreateUser(user *v1.UserSpec) (*IdentityUser, error) {
	// prepare request url
	url := "http://" + s.config.host + ":" + strconv.Itoa(s.config.port) + "/users"

	// prepare request body
	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	// prepare request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// set authorization header with token
	req.Header.Set("Authorization", "Bearer "+s.token)

	// set content type header
	req.Header.Set("Content-Type", "application/json")

	// make REST API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// close the response body
	defer resp.Body.Close()

	// read response body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse response body
	var userResponse IdentityUser
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	// return the IdentityUser object
	return &userResponse, nil
}

// GetUser retrieves the user with the given ID from external identity app using REST API call.
func (s *IdentityService) GetUser(userID string) (*IdentityUser, error) {
	// prepare request URL
	url := "http://" + s.config.host + ":" + strconv.Itoa(s.config.port) + "/users/" + userID

	// create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// set authorization header with token
	req.Header.Set("Authorization", "Bearer "+s.token)

	// set accept header to JSON
	req.Header.Set("Accept", "application/json")

	// make REST API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	var userResponse IdentityUser
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	// return the user object
	return &userResponse, nil
}

func (s *IdentityService) DeleteUser(userID string) error {
	// prepare request URL
	url := "http://" + s.config.host + ":" + strconv.Itoa(s.config.port) + "/users/" + userID

	// create request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	// set authorization header with token
	req.Header.Set("Authorization", "Bearer "+s.token)

	// make REST API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// close the response body
	defer resp.Body.Close()

	return nil
}

func (s *IdentityService) UpdateUser(userID string, user *v1.UserSpec) (*IdentityUser, error) {
	// prepare request URL
	url := "http://" + s.config.host + ":" + strconv.Itoa(s.config.port) + "/users/" + userID

	// prepare request body
	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	// prepare request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// set authorization header with token
	req.Header.Set("Authorization", "Bearer "+s.token)

	// set content type header
	req.Header.Set("Content-Type", "application/json")

	// make REST API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// close the response body
	defer resp.Body.Close()

	// read response body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	var userResponse IdentityUser
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	// return the user object
	return &userResponse, nil
}
