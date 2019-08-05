package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// CAI is the SAP Conversational AI interface.
type CAI struct {
	Token string
}

// Intent defines the intent of a message.
type Intent struct {
	Slug     string
	FullName string
	Lat      float64
	Lng      float64
	Address  string
	Product  string
	Mass     float64
	Number   uint
	Dollars  float64
}

// NewCAI initializes a new CAI interface.
func NewCAI(token string) *CAI {
	return &CAI{Token: token}
}

// Intent returns the intent of a message.
func (cai *CAI) Intent(message string) (*Intent, error) {
	type IntentResult struct {
		Results struct {
			Intents []struct {
				Slug string `json:"slug"`
			} `json:"intents"`
			Entities map[string][]struct {
				FullName  string  `json:"fullname"`
				Lat       float64 `json:"lat"`
				Lng       float64 `json:"lng"`
				Scalar    float64 `json:"scalar"`
				Value     string  `json:"value"`
				Grams     float64 `json:"grams"`
				Formatted string  `json:"formatted"`
				Dollars   float64 `json:"dollars"`
			} `json:"entities"`
		} `json:"results"`
	}

	payload := url.Values{"text": {message}, "language": {"en"}}
	req, err := http.NewRequest("POST", "https://api.cai.tools.sap/v2/request",
		strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", cai.Token))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result IntentResult
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Results.Intents) < 1 {
		return nil, errors.New("No intent")
	}

	intent := &Intent{Slug: result.Results.Intents[0].Slug}
	if value, ok := result.Results.Entities["person"]; ok {
		intent.FullName = value[0].FullName
	}
	if values, ok := result.Results.Entities["location"]; ok {
		intent.Lat = values[0].Lat
		intent.Lng = values[0].Lng
		intent.Address = values[0].Formatted
	}
	if values, ok := result.Results.Entities["product"]; ok {
		intent.Product = values[0].Value
	}
	if values, ok := result.Results.Entities["mass"]; ok {
		intent.Mass = values[0].Grams
	}
	if values, ok := result.Results.Entities["number"]; ok {
		intent.Number = uint(values[0].Scalar)
	}
	if values, ok := result.Results.Entities["money"]; ok {
		intent.Dollars = values[0].Dollars
	}

	return intent, nil
}
