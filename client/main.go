package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/rosti-cz/server_lobby/server"
)

// Encapsulation of Lobby's client code
type LobbyClient struct {
	Proto string
	Host  string
	Port  uint
	Token string
}

func (l *LobbyClient) init() {
	if len(l.Proto) == 0 {
		l.Host = "http"
	}
	if len(l.Host) == 0 {
		l.Host = "localhost"
	}
	if l.Port == 0 {
		l.Port = 1313
	}
}

// calls the backend API with given method, path and request body and returns status code, response body and error if there is any.
// Method can be GET, POST or DELETE.
// Path should start with / and it can contain query parameters too.
func (l *LobbyClient) call(method, path, body string) (uint, string, error) {
	client := resty.New().R()

	if len(l.Token) != 0 {
		client = client.SetHeader("Authorization", fmt.Sprintf("Token %s", l.Token))
	}

	if strings.ToUpper(method) == "GET" {
		resp, err := client.Get(fmt.Sprintf("%s://%s:%d%s", l.Proto, l.Host, l.Port, path))
		if err != nil {
			return 0, "", err
		}
		return uint(resp.StatusCode()), string(resp.Body()), nil
	} else if strings.ToUpper(method) == "POST" {
		resp, err := client.SetBody(body).Post(fmt.Sprintf("%s://%s:%d%s", l.Proto, l.Host, l.Port, path))
		if err != nil {
			return 0, "", err
		}
		return uint(resp.StatusCode()), string(resp.Body()), nil
	} else if strings.ToUpper(method) == "DELETE" {
		resp, err := client.SetBody(body).Delete(fmt.Sprintf("%s://%s:%d%s", l.Proto, l.Host, l.Port, path))
		if err != nil {
			return 0, "", err
		}
		return uint(resp.StatusCode()), string(resp.Body()), nil
	} else {
		return 0, "", errors.New("unsupported method")
	}

}

// Returns discovery object of local machine
func (l *LobbyClient) GetDiscovery() (server.Discovery, error) {
	l.init()

	var discovery server.Discovery

	path := "/v1/discovery"
	method := "GET"

	status, body, err := l.call(method, path, "")
	if err != nil {
		return discovery, err
	}
	if status != 200 {
		return discovery, fmt.Errorf("non-200 response: %s", body)
	}

	err = json.Unmarshal([]byte(body), &discovery)
	if err != nil {
		return discovery, fmt.Errorf("response parsing error: %v", err)
	}

	return discovery, nil
}

// Returns all registered discovery packets
func (l *LobbyClient) GetDiscoveries() ([]server.Discovery, error) {
	l.init()

	path := "/v1/discoveries"
	method := "GET"

	var discoveries []server.Discovery

	status, body, err := l.call(method, path, "")
	if err != nil {
		return discoveries, err
	}
	if status != 200 {
		return discoveries, fmt.Errorf("non-200 response: %s", body)
	}

	err = json.Unmarshal([]byte(body), &discoveries)
	if err != nil {
		return discoveries, fmt.Errorf("response parsing error: %v", err)
	}

	return discoveries, nil
}

// Find discoveries by their labels
func (l *LobbyClient) FindByLabels(labels server.Labels) (server.Discoveries, error) {
	l.init()

	path := fmt.Sprintf("/v1/discoveries?labels=%s", strings.Join(labels.StringSlice(), ","))
	method := "GET"

	var discoveries server.Discoveries

	status, body, err := l.call(method, path, "")
	if err != nil {
		return discoveries, err
	}
	if status != 200 {
		return discoveries, fmt.Errorf("non-200 response: %s", body)
	}

	err = json.Unmarshal([]byte(body), &discoveries)
	if err != nil {
		return discoveries, fmt.Errorf("response parsing error: %v", err)
	}

	return discoveries, nil
}

// Adds runtime labels for the local machine
func (l *LobbyClient) AddLabels(labels server.Labels) error {
	l.init()

	path := "/v1/labels"
	method := "POST"

	status, body, err := l.call(method, path, strings.Join(labels.StringSlice(), "\n"))
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("non-200 response: %s", body)
	}

	return nil
}

// Removes runtime labels of the local machine
func (l *LobbyClient) DeleteLabels(labels server.Labels) error {
	l.init()

	path := "/v1/labels"
	method := "DELETE"

	status, body, err := l.call(method, path, strings.Join(labels.StringSlice(), "\n"))
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("non-200 response: %s", body)
	}

	return nil
}
