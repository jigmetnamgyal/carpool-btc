package helpers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Payload struct {
	JsonRPC float64       `json:"json_rpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func SendRequest(payload Payload) (map[string]interface{}, error) {
	url := "http://localhost:18332"

	jsonData, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, err
	}

	username := os.Getenv("BITCOIN_CORE_USERNAME")
	password := os.Getenv("BITCOIN_CORE_PASSWORD")

	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp)
		return nil, errors.New("server returned non-200 status")
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		return nil, err
	}

	if jsonValues, ok := data.(map[string]interface{}); ok {
		return jsonValues, nil
	}

	return nil, fmt.Errorf("unexpected type")
}
