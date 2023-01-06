package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	SyncRequestType  = "sync_request"
	SyncResponseType = "sync_response"
)

// representation of connected saiP2p connected nodes
type SaiP2pNode struct {
	Address string `json:"address"`
}

type SyncRequest struct {
	Type    string `json:"type"`
	From    int    `json:"block_number_from"`
	To      int    `json:"block_number_to"`
	Address string `json:"address"`
}

type SyncResponse struct {
	Type  string `json:"type"`
	Error error  `json:"error"`
	Link  string `json:"link"`
}

// send direct get block message to connected nodes
func SendDirectGetBlockMsg(node string, req *SyncRequest, saiP2pAddress string) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("chain - sendDirectGetBlockMsg - marshal request : %w", err)
	}

	param := url.Values{}
	param.Add("message", string(data))
	param.Add("node", node)

	postRequest, err := http.NewRequest("POST", saiP2pAddress+"/Send_message_to", strings.NewReader(param.Encode()))
	if err != nil {
		return fmt.Errorf("chain - sendDirectGetBlockMsg - create post request : %w", err)
	}

	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(postRequest)
	if err != nil {
		return fmt.Errorf("chain - sendDirectGetBlockMsg - send post request : %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("chain - sendDirectGetBlockMsg - send post request wrong response status code : %d", resp.StatusCode)
	}

	return nil

}
