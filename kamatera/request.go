package kamatera

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

func request(provider *ProviderConfig, method string, path string, body interface{}) (interface{}, error) {
	if provider == nil {
		return nil, noProviderErr
	}

	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, _ := http.NewRequest(method, fmt.Sprintf("%s/%s", provider.ApiUrl, path), buf)
	req.Header.Add("AuthClientId", provider.ApiClientID)
	req.Header.Add("AuthSecret", provider.ApiSecret)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	client := cleanhttp.DefaultClient()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("bad status code from Kamatera API: %d", res.StatusCode)
		} else {
			return nil, fmt.Errorf("invalid response from Kamatera API: %+v", result)
		}
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error response from Kamatera API (%d): %+v", res.StatusCode, result)
	}
	return result, nil
}

func postServerConfigure(provider *ProviderConfig, postValues configureServerPostValues) error {
	if provider == nil {
		return noProviderErr
	}

	result, err := request(provider, "POST", "server/configure", postValues)
	if err != nil {
		return err
	}

	commandIds := result.([]interface{})
	_, err = waitCommand(provider, commandIds[0].(string))
	return err
}

func serverChangePassword(provider *ProviderConfig, internalServerID string, password string) error {
	result, err := request(provider, "POST", "service/server/password", changePasswordServerPostValues{ID: internalServerID, Password: password})
	if err != nil {
		return err
	}

	commandIds := result.([]interface{})
	_, err = waitCommand(provider, commandIds[0].(string))

	return err
}

func renameServer(provider *ProviderConfig, internalServerID string, name string) error {
	result, err := request(
		provider,
		"POST",
		fmt.Sprintf("service/server/rename"),
		renameServerPostValues{ID: internalServerID, NewName: name},
	)
	if err != nil {
		return err
	}

	commandIds := result.([]interface{})
	_, err = waitCommand(provider, commandIds[0].(string))

	return err
}

type diskOperation struct {
	add    []float64
	remove []int           // index
	update map[int]float64 // map[index]newValue
}

func changeDisks(provider *ProviderConfig, id string, operation diskOperation) error {
	switch {
	case len(operation.add) > 0:
		for _, v := range operation.add {
			result, err := request(
				provider,
				"POST",
				"server/disk",
				changeDisksPostValues{
					ID:  id,
					Add: fmt.Sprintf("%fgb", v),
				},
			)
			if err != nil {
				return err
			}

			commandIds := result.([]interface{})
			_, err = waitCommand(provider, commandIds[0].(string))
		}
	case len(operation.remove) > 0:
		for _, v := range operation.remove {
			result, err := request(
				provider,
				"POST",
				"server/disk",
				changeDisksPostValues{
					ID:     id,
					Remove: fmt.Sprint(v),
				},
			)
			if err != nil {
				return err
			}

			commandIds := result.([]interface{})
			_, err = waitCommand(provider, commandIds[0].(string))
		}
	case len(operation.update) > 0:
		for key, val := range operation.update {
			result, err := request(
				provider,
				"POST",
				"server/disk",
				changeDisksPostValues{
					ID:     id,
					Resize: fmt.Sprint(key),
					Size:   fmt.Sprintf("%fgb", val),
				},
			)
			if err != nil {
				return err
			}

			commandIds := result.([]interface{})
			_, err = waitCommand(provider, commandIds[0].(string))
		}
	}

	return nil
}

func waitCommand(provider *ProviderConfig, commandID string) (map[string]interface{}, error) {
	if provider == nil {
		return nil, noProviderErr
	}

	startTime := time.Now()
	time.Sleep(2 * time.Second)

	for {
		if startTime.Add(40*time.Minute).Sub(time.Now()) < 0 {
			return nil, errors.New("timeout waiting for Kamatera command to complete")
		}

		time.Sleep(2 * time.Second)

		result, e := request(provider, "GET", fmt.Sprintf("service/queue?id=%s", commandID), nil)
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
					return nil, fmt.Errorf("kamatera command failed: %s", log)
				} else {
					return nil, fmt.Errorf("kamatera command failed: %v", command)
				}
			}
		}
	}
}
