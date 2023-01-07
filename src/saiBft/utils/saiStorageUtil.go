package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Database struct {
	url                  string
	email                string
	password             string
	duplicateRequests    bool
	duplicateRequestsUrl string
	duplicateRequestsCh  chan *bytes.Buffer
}

func Storage(Url string, Email string, Password string, duplicateRequests bool, duplicateRequestsUrl string, duplicateCh chan *bytes.Buffer) Database {
	return Database{
		url:                  Url,
		email:                Email,
		password:             Password,
		duplicateRequests:    duplicateRequests,
		duplicateRequestsUrl: duplicateRequestsUrl,
		duplicateRequestsCh:  duplicateCh,
	}
}

type StorageRequest struct {
	token      string
	collection string
	options    interface{}
	criteria   interface{}
	data       interface{}
}

func (s StorageRequest) toJson() ([]byte, error) {
	jsonObj := bson.M{"collection": s.collection}

	if s.data != nil {
		jsonObj["data"] = s.data
	}

	if s.criteria != nil {
		jsonObj["select"] = s.criteria
	}

	if s.options != nil {
		jsonObj["options"] = s.options
	}

	return json.Marshal(jsonObj)
}

func (db Database) Get(collectionName string, criteria interface{}, options interface{}, token string) (error, []byte) {
	request := StorageRequest{collection: collectionName, criteria: criteria, options: options}
	return db.makeRequest("get", request, token)
}

func (db Database) Put(collectionName string, data interface{}, token string) (error, []byte) {
	request := StorageRequest{collection: collectionName, data: data}
	return db.makeRequest("save", request, token)
}

func (db Database) Update(collectionName string, criteria interface{}, data interface{}, token string) (error, []byte) {
	request := StorageRequest{collection: collectionName, criteria: criteria, data: data}
	return db.makeRequest("update", request, token)
}

func (db Database) Upsert(collectionName string, criteria interface{}, data interface{}, token string) (error, []byte) {
	request := StorageRequest{collection: collectionName, criteria: criteria, data: data}
	return db.makeRequest("upsert", request, token)
}

func (db Database) Remove(collectionName string, criteria interface{}, token string) (error, []byte) {
	request := StorageRequest{collection: collectionName, criteria: criteria}
	return db.makeRequest("remove", request, token)
}

func (db Database) makeRequest(method string, request StorageRequest, token string) (error, []byte) {
	jsonStr, jsonErr := request.toJson()

	//fmt.Println("Storage request:", string(jsonStr))

	if jsonErr != nil {
		fmt.Println("Database request error: ", jsonErr)
		return jsonErr, []byte("")
	}
	if db.duplicateRequests {
		if method == "save" || method == "upsert" || method == "update" {
			go func() {
				db.duplicateRequestsCh <- bytes.NewBuffer(jsonStr)
			}()
		}
	}

	return Send(db.url+"/"+method, bytes.NewBuffer(jsonStr), token)
}

func Send(url string, data io.Reader, token string) (error, []byte) {
	req, err := http.NewRequest("POST", url, data)

	if err != nil {
		fmt.Println("Database error: ", err)
		return err, []byte("")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Database error: ", err)
		return err, []byte("")
	}

	defer resp.Body.Close()
	_ = time.AfterFunc(5*time.Second, func() {
		resp.Body.Close()
	})
	body, _ := ioutil.ReadAll(resp.Body)
	return nil, body
}
