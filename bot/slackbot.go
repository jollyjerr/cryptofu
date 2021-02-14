package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	httpClient = http.Client{
		Timeout: time.Second * 30,
	}
)

// SendSlackFinancials is for buy and sell logging only
func SendSlackFinancials(message interface{}) error {
	_, err := post(os.Getenv("SLACK_URL"), message)
	if err != nil {
		return err
	}
	return nil
}

// SendSlackLogging sends general logs
func SendSlackLogging(message interface{}) error {
	_, err := post(os.Getenv("SLACK_URL"), message)
	if err != nil {
		return err
	}
	return nil
}

func post(url string, body interface{}) (*http.Response, error) {
	marshaledBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	message := map[string]interface{}{
		"text": string(marshaledBody),
	}
	marshaledBody, err = json.Marshal(message)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshaledBody))
	if err != nil {
		return nil, err
	}

	addStandardHeaders(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	return resp, nil
}

func addStandardHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "cryptofu")
	req.Header.Set("Accept", "application/json")
}
