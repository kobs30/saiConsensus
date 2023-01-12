package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// extract result from saiStorage service (crop raw data)
// {"result":[........] --> [.....]}
func ExtractResult(input []byte) ([]byte, error) {
	_, after, found := bytes.Cut(input, []byte(":"))
	if !found {
		return nil, errors.New("wrong result!")
	}

	result := bytes.TrimSuffix(after, []byte("}"))
	return result, nil

}

type IP struct {
	Query string `json:"query"`
}

func GetOutboundIP() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	var ip IP
	json.Unmarshal(body, &ip)

	return ip.Query
}

func UniqueStrings(slice ...[]string) []string {
	uniqueMap := map[string]bool{}

	for _, intSlice := range slice {
		for _, s := range intSlice {
			uniqueMap[s] = true
		}
	}

	// Create a slice with the capacity of unique items
	// This capacity make appending flow much more efficient
	result := make([]string, 0, len(uniqueMap))

	for key := range uniqueMap {
		result = append(result, key)
	}

	return result
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SendHttpRequest(url string, payload []byte) (interface{}, bool) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))

	if err != nil {
		fmt.Println("Call VM error: ", err)
		return nil, false
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Call VM error: ", err)
		return nil, false
	}

	defer resp.Body.Close()
	_ = time.AfterFunc(5*time.Second, func() {
		resp.Body.Close()
	})
	body, _ := ioutil.ReadAll(resp.Body)

	return body, true
}
