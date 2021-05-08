package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/kamatera/terraform-provider-kamatera/kamatera"
)

func request(provider kamatera.ProviderConfiguration, method string, path string, body interface{}) (interface{}, error) {
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
