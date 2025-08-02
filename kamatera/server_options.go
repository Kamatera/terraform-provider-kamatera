package kamatera

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
)

var (
	serverOptions         gjson.Result
	serverOptionsErr      error
	loadServerOptionsOnce sync.Once
)

func _loadServerOptions() (gjson.Result, error) {
	resp, err := http.Get("https://console.kamatera.com/info/calculator.js.php")
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to download server options: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to read response body: %w", err)
	}
	parts := strings.Split(string(body), "'")
	if len(parts) < 3 {
		return gjson.Result{}, fmt.Errorf("unexpected content format")
	}
	jsonStr := strings.Join(parts[1:len(parts)-1], "'")
	if !gjson.Valid(jsonStr) {
		return gjson.Result{}, fmt.Errorf("invalid JSON format in response")
	}
	return gjson.Parse(jsonStr), nil
}

func loadServerOptions() error {
	loadServerOptionsOnce.Do(func() {
		serverOptions, serverOptionsErr = _loadServerOptions()
	})
	return serverOptionsErr
}

func serverOptionsValidateDatacenter(datacenterId string) error {
	found := false
	for key, _ := range serverOptions.Map() {
		if key == fmt.Sprintf("netPck.%s", datacenterId) {
			found = true
			break
		}
	}
	for _, os := range serverOptions.Get("os").Array() {
		for _, d := range os.Get("datacenters").Array() {
			if d.String() == datacenterId {
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("unsupported datacenter ID: %s", datacenterId)
	}
	return nil
}

func serverOptionsValidateCpu(cpu string) error {
	found := false
	for _, option := range serverOptions.Get("cpu.0.options").Array() {
		if option.Get("value").String() == cpu {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported CPU: %s", cpu)
	}
	return nil
}

func serverOptionsValidateDiskSizeGB(diskSizeGB int) error {
	found := false
	for _, option := range serverOptions.Get("diskGB.0.options").Array() {
		if option.Get("value").Int() == int64(diskSizeGB) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported disk size: %d GB", diskSizeGB)
	}
	return nil
}

func serverOptionsValidateMonthlyTrafficPackage(datacenterId string, monthlyTrafficPackage string) error {
	found := false
	for _, option := range serverOptions.Get(fmt.Sprintf("netPck\\.%s.0.options", datacenterId)).Array() {
		if option.Get("value").String() == monthlyTrafficPackage {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported monthly traffic package for %s datacenter: %s", datacenterId, monthlyTrafficPackage)
	}
	return nil
}

func serverOptionsValidateRamMB(cpuType string, ramMB int) error {
	found := false
	for _, option := range serverOptions.Get(fmt.Sprintf("ramMB\\.%s.0.options", cpuType)).Array() {
		if option.Get("value").Int() == int64(ramMB) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported RAM size for CPU type %s: %d MB", cpuType, ramMB)
	}
	return nil
}
