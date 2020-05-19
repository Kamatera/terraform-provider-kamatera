package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func kamateraRequest(provider ProviderConfiguration, method string, path string, body interface{}) (interface{}, error) {
	buf := new(bytes.Buffer)
	if body != nil {
		if e := json.NewEncoder(buf).Encode(body); e != nil {
			return nil, e
		}
	}
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/%s", provider.ApiUrl, path), buf)
	req.Header.Add("AuthClientId", provider.ApiClientID)
	req.Header.Add("AuthSecret", provider.ApiSecret)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	var result interface{}
	e = json.NewDecoder(res.Body).Decode(&result)
	if e != nil {
		if res.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("invalid response from Kamatera API: %d", res.StatusCode))
		} else {
			return nil, errors.New(fmt.Sprintf("invalid response from Kamatera API: %v", result))
		}
	}
	if res.StatusCode != 200 {
		switch val := result.(type) {
		case map[string]interface{}:
			if resultMessage, hasMessage := val["message"]; hasMessage {
				return nil, errors.New(fmt.Sprintf("error response from Kamatera API (%d): %v", res.StatusCode, resultMessage))
			}
		}
		return nil, errors.New(fmt.Sprintf("error response from Kamatera API (%d): %v", res.StatusCode, result))
	}
	return result, nil
}

func waitCommand(provider ProviderConfiguration, commandId string) (map[string]interface{}, error) {
	startTime := time.Now()
	time.Sleep(2 * time.Second)
	for {
		if startTime.Add(2400 * time.Second).Sub(time.Now()) < 0 {
			return nil, errors.New("timeout waiting for Kamatera command to complete")
		}
		time.Sleep(2 * time.Second)
		result, e := kamateraRequest(provider, "GET", fmt.Sprintf("service/queue?id=%s", commandId), nil)
		if e != nil {
			return nil, e
		}
		commands := result.([]interface{})
		if len(commands) != 1 {
			return nil, errors.New("invalid response from Kamatera queue API: invalid number of command responses")
		}
		command := commands[0].(map[string]interface{})
		status, hasStatus := command["status"]
		if hasStatus {
			switch status.(string) {
			case "complete":
				return command, nil
			case "error":
				log, hasLog := command["log"]
				if hasLog {
					return nil, errors.New(fmt.Sprintf("Kamatera command failed: %s", log))
				} else {
					return nil, errors.New(fmt.Sprintf("Kamatera command failed: %v", command))
				}
			}
		}
	}
}
