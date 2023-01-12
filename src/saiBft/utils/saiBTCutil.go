package utils

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// send and handle response from saiBTC service
func SendRequest(address string, requestBody io.Reader) ([]byte, error) {
	keysRequest, err := http.NewRequest("POST", address, requestBody)
	if err != nil {
		return nil, err
	}
	keysRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(keysRequest)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
