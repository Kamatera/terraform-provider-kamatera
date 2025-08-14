package kamatera

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestServerOptionsValidateDatacenter(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, datacenter := range []string{"EU", "US", "CA-TR"} {
		assert.NoError(t, serverOptionsValidateDatacenter(datacenter), "unexpected error for: %s", datacenter)
	}
	for _, datacenter := range []string{"EU1", "US2", "ASIA3", ""} {
		err := serverOptionsValidateDatacenter(datacenter)
		assert.NotNil(t, err, "expected error, got nil for: %s", datacenter)
		errMsg := err.Error()
		assert.Contains(t, errMsg, fmt.Sprintf("Unsupported datacenter ID: %s\nSupported datacenter IDs: ", datacenter))
		supportedDatacenterIds := strings.Split(strings.Split(errMsg, "Supported datacenter IDs: ")[1], ", ")
		assert.Len(t, supportedDatacenterIds, 26, "unexpected number of supported datacenter IDs: %d", len(supportedDatacenterIds))
		assert.Contains(t, supportedDatacenterIds, "EU", "expected 'EU' in supported datacenter IDs")
		assert.Contains(t, supportedDatacenterIds, "US", "expected 'US' in supported datacenter IDs")
		assert.Contains(t, supportedDatacenterIds, "CA-TR", "expected 'CA-TR' in supported datacenter IDs")
	}
}

func TestServerOptionsValidateCpu(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, cpu := range []string{"1A", "2B", "4T", "8D"} {
		assert.NoError(t, serverOptionsValidateCpu(cpu))
	}
	for _, cpu := range []string{"11A", "33B", "9999T", "123D"} {
		cpuSplit := strings.Split(cpu, "")
		cpuType := cpuSplit[len(cpuSplit)-1]
		cpuNum := strings.Join(cpuSplit[:len(cpuSplit)-1], "")
		err := serverOptionsValidateCpu(cpu)
		assert.NotNil(t, err, "expected error, got nil for: %s", cpu)
		errMsg := err.Error()
		assert.Contains(
			t,
			errMsg,
			fmt.Sprintf(
				"unsupported number of cores for CPU type %s: %s\nsupported number of cores for this CPU type: ",
				cpuType, cpuNum,
			),
		)
		supportedCpuNums := strings.Split(strings.Split(errMsg, "supported number of cores for this CPU type: ")[1], ", ")
		if cpuType == "A" {
			assert.Len(t, supportedCpuNums, 11, "unexpected number of supported CPUs for type A: %s", strings.Join(supportedCpuNums, ", "))
		} else if cpuType == "B" {
			assert.Len(t, supportedCpuNums, 19, "unexpected number of supported CPUs for type B: %s", strings.Join(supportedCpuNums, ", "))
		} else if cpuType == "T" {
			assert.Len(t, supportedCpuNums, 19, "unexpected number of supported CPUs for type T: %s", strings.Join(supportedCpuNums, ", "))
		} else if cpuType == "D" {
			assert.Len(t, supportedCpuNums, 13, "unexpected number of supported CPUs for type D: %s", strings.Join(supportedCpuNums, ", "))
		}
	}
}

func TestServerOptionsValidateDiskSizeGB(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, size := range []int{5, 10, 15, 2000, 3000} {
		assert.NoError(t, serverOptionsValidateDiskSizeGB(size))
	}
	for _, size := range []int{-1, 0, 3, 2002} {
		err := serverOptionsValidateDiskSizeGB(size)
		assert.NotNil(t, err)
		errMsg := err.Error()
		assert.Contains(t, errMsg, fmt.Sprintf("unsupported disk size: %d GB\nsupported disk sizes: ", size))
		supportedDiskSizes := strings.Split(strings.Split(errMsg, "supported disk sizes: ")[1], ", ")
		assert.Len(t, supportedDiskSizes, 27, "unexpected number of supported disk sizes: %s", strings.Join(supportedDiskSizes, ", "))
		assert.Contains(t, supportedDiskSizes, "5")
		assert.Contains(t, supportedDiskSizes, "15")
		assert.Contains(t, supportedDiskSizes, "30")
	}
}

func TestServerOptionsValidateMonthlyTrafficPackage(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, mtp := range []string{"t5000", "b50"} {
		assert.NoError(t, serverOptionsValidateMonthlyTrafficPackage("EU", mtp), "Monthly traffic package validation failed for: %v", mtp)
	}
	for _, mtp := range []string{"t5001", "d", ""} {
		err := serverOptionsValidateMonthlyTrafficPackage("EU", mtp)
		assert.NotNil(t, err)
		errMsg := err.Error()
		assert.Contains(t, errMsg, fmt.Sprintf(
			"unsupported monthly traffic package for EU datacenter: %s\nsupported packages for this datacenter: ", mtp,
		))
		supportedPackages := strings.Split(strings.Split(errMsg, "supported packages for this datacenter: ")[1], ", ")
		assert.Len(t, supportedPackages, 2, "unexpected number of supported packages: %s", strings.Join(supportedPackages, ", "))
		assert.Contains(t, supportedPackages, "t5000")
		assert.Contains(t, supportedPackages, "b50")
	}
}

func TestServerOptionsValidateRamMb(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, ram := range []int{512, 1024, 2048, 4096, 8192} {
		assert.NoError(t, serverOptionsValidateRamMB("D", ram), "RAM validation failed for: %d MB", ram)
	}
	for _, ram := range []int{-1, 0, 123, 8096} {
		err := serverOptionsValidateRamMB("D", ram)
		assert.NotNil(t, err)
		errMsg := err.Error()
		assert.Contains(t, errMsg, fmt.Sprintf("unsupported RAM size for CPU type D: %d MB\nsupported RAM sizes for this CPU type: ", ram))
		supportedRamSizes := strings.Split(strings.Split(errMsg, "supported RAM sizes for this CPU type: ")[1], ", ")
		assert.Len(t, supportedRamSizes, 23, "unexpected number of supported RAM sizes: %s", strings.Join(supportedRamSizes, ", "))
		assert.Contains(t, supportedRamSizes, "512")
		assert.Contains(t, supportedRamSizes, "1024")
		assert.Contains(t, supportedRamSizes, "2048")
	}
}
