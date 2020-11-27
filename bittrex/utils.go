package bittrex

import (
	"encoding/json"
	"io/ioutil"
)

func getCandleResponse(path string) ([]CandleResponse, error) {
	result := []CandleResponse{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(file), &result)
	return result, err
}
