package kamatera

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerOptionsValidateDatacenter(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, datacenter := range []string{"EU", "US", "CA-TR"} {
		assert.NoError(t, serverOptionsValidateDatacenter(datacenter), "Datacenter validation failed for: %s", datacenter)
	}
	for _, datacenter := range []string{"EU1", "US2", "ASIA3", ""} {
		assert.Error(t, serverOptionsValidateDatacenter(datacenter), "Datacenter validation should fail for: %s", datacenter)
	}
}

func TestServerOptionsValidateCpu(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, cpu := range []string{"1A", "2B", "4T", "8D"} {
		assert.NoError(t, serverOptionsValidateCpu(cpu), "CPU validation failed for: %s", cpu)
	}
	for _, cpu := range []string{"1C", "3B", "9999T", "D", ""} {
		assert.Error(t, serverOptionsValidateCpu(cpu), "CPU validation should fail for: %s", cpu)
	}
}

func TestServerOptionsValidateDiskSizeGB(t *testing.T) {
	err := loadServerOptions()
	if err != nil {
		t.Fatalf("Failed to load server options: %v", err)
	}
	for _, size := range []int{5, 10, 15, 2000, 3000} {
		assert.NoError(t, serverOptionsValidateDiskSizeGB(size), "Disk size validation failed for: %d GB", size)
	}
	for _, size := range []int{-1, 0, 3, 2002} {
		assert.Error(t, serverOptionsValidateDiskSizeGB(size), "Disk size validation should fail for: %d GB", size)
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
		assert.Error(t, serverOptionsValidateMonthlyTrafficPackage("EU", mtp), "Monthly traffic package validation should fail for: %v", mtp)
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
		assert.Error(t, serverOptionsValidateRamMB("D", ram), "RAM validation should fail for: %d MB", ram)
	}
}
