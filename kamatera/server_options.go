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

func commaJoinMapKeys(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return strings.Join(keys, ", ")
}

func serverOptionsValidateDatacenter(datacenterId string) error {
	found := false
	validDatacenters := make(map[string]struct{})
	for key, _ := range serverOptions.Map() {
		if key == fmt.Sprintf("netPck.%s", datacenterId) {
			found = true
			break
		} else if strings.HasPrefix(key, "netPck.") {
			validDatacenters[strings.Replace(key, "netPck.", "", 1)] = struct{}{}
		}
	}
	for _, os := range serverOptions.Get("os").Array() {
		for _, d := range os.Get("datacenters").Array() {
			if d.String() == datacenterId {
				found = true
				break
			} else {
				validDatacenters[d.String()] = struct{}{}
			}
		}
	}
	if !found {
		return fmt.Errorf(
			"Unsupported datacenter ID: %s\nSupported datacenter IDs: %s",
			datacenterId, commaJoinMapKeys(validDatacenters),
		)
	}
	return nil
}

func serverOptionsValidateCpu(cpu string) error {
	found := false
	cpuSplit := strings.Split(cpu, "")
	if len(cpuSplit) < 2 {
		// this should never happen, because we create the cpu string in the correct format
		return fmt.Errorf("invalid CPU format: %s", cpu)
	}
	cpuType := cpuSplit[len(cpuSplit)-1]
	if cpuType != "A" && cpuType != "B" && cpuType != "T" && cpuType != "D" {
		// this should never happen, because we validate this in the server resource field
		return fmt.Errorf("invalid CPU type: %s", cpuType)
	}
	cpuNum := strings.Join(cpuSplit[:len(cpuSplit)-1], "")
	validCpuNums := make(map[string]struct{})
	for _, option := range serverOptions.Get("cpu.0.options").Array() {
		if option.Get("value").String() == cpu {
			found = true
			break
		} else {
			optCpuSplit := strings.Split(option.Get("value").String(), "")
			if len(optCpuSplit) >= 2 {
				optCpuType := optCpuSplit[len(optCpuSplit)-1]
				optCpuNum := strings.Join(optCpuSplit[:len(optCpuSplit)-1], "")
				if optCpuType == cpuType {
					validCpuNums[optCpuNum] = struct{}{}
				}
			}
		}
	}
	if !found {
		return fmt.Errorf(
			"unsupported number of cores for CPU type %s: %s\nsupported number of cores for this CPU type: %s",
			cpuType,
			cpuNum,
			commaJoinMapKeys(validCpuNums),
		)
	}
	return nil
}

func serverOptionsValidateDiskSizeGB(diskSizeGB int) error {
	found := false
	validDiskSizes := make(map[string]struct{})
	for _, option := range serverOptions.Get("diskGB.0.options").Array() {
		if option.Get("value").Int() == int64(diskSizeGB) {
			found = true
			break
		} else {
			validDiskSizes[option.Get("value").String()] = struct{}{}
		}
	}
	if !found {
		return fmt.Errorf(
			"unsupported disk size: %d GB\nsupported disk sizes: %s",
			diskSizeGB,
			commaJoinMapKeys(validDiskSizes),
		)
	}
	return nil
}

func serverOptionsValidateMonthlyTrafficPackage(datacenterId string, monthlyTrafficPackage string) error {
	found := false
	validPackages := make(map[string]struct{})
	for _, option := range serverOptions.Get(fmt.Sprintf("netPck\\.%s.0.options", datacenterId)).Array() {
		if option.Get("value").String() == monthlyTrafficPackage {
			found = true
			break
		} else {
			validPackages[option.Get("value").String()] = struct{}{}
		}
	}
	if !found {
		return fmt.Errorf(
			"unsupported monthly traffic package for %s datacenter: %s\nsupported packages for this datacenter: %s",
			datacenterId,
			monthlyTrafficPackage,
			commaJoinMapKeys(validPackages),
		)
	}
	return nil
}

func serverOptionsValidateRamMB(cpuType string, ramMB int) error {
	found := false
	validRamSizes := make(map[string]struct{})
	for _, option := range serverOptions.Get(fmt.Sprintf("ramMB\\.%s.0.options", cpuType)).Array() {
		if option.Get("value").Int() == int64(ramMB) {
			found = true
			break
		} else {
			validRamSizes[option.Get("value").String()] = struct{}{}
		}
	}
	if !found {
		return fmt.Errorf(
			"unsupported RAM size for CPU type %s: %d MB\nsupported RAM sizes for this CPU type: %s",
			cpuType,
			ramMB,
			commaJoinMapKeys(validRamSizes),
		)
	}
	return nil
}
