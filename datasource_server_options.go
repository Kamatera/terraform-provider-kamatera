package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kamatera/terraform-provider-kamatera/disk/helper"
)

func dataSourceServerOptions() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceServerOptionsRead,

		Schema: map[string]*schema.Schema{
			"datacenter_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"cpu_cores": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"ram_mb": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"billing_cycle": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "hourly",
			},
			"monthly_traffic_package": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			// TODO: Remove deprecated field.
			"disk_size_gb": &schema.Schema{
				Type:       schema.TypeFloat,
				Optional:   true,
				Deprecated: "Use disks instead",
			},
			// TODO: Remove deprecated field.
			"extra_disk_sizes_gb": &schema.Schema{
				Type:       schema.TypeList,
				Elem:       &schema.Schema{Type: schema.TypeFloat},
				MaxItems:   3,
				Optional:   true,
				Deprecated: "Use disks instead",
			},
			"disks": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
				MaxItems: 4,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return []interface{}{"10GB"}, nil
				},
			},
		},
	}
}

func DataSourceServerOptionsRead(d *schema.ResourceData, m interface{}) error {
	provider := m.(*ProviderConfiguration)
	result, e := kamateraRequest(*provider, "GET", fmt.Sprintf("service/server?capabilities=1&datacenter=%s", d.Get("datacenter_id").(string)), nil)
	if e != nil {
		d.SetId("")
		return e
	}
	serverOptions := result.(map[string]interface{})
	cpuTypes := serverOptions["cpuTypes"].([]interface{})
	monthlyTrafficPackage := serverOptions["monthlyTrafficPackage"].(map[string]interface{})
	diskSizeGB := serverOptions["diskSizeGB"].([]float64)
	defaultMonthlyTrafficPackage := serverOptions["defaultMonthlyTrafficPackage"].(string)
	var availableCpuTypes []string
	validCpuType := false
	for _, cpuType := range cpuTypes {
		cpuType := cpuType.(map[string]interface{})
		cpuTypeCores := cpuType["cpuCores"].([]interface{})
		cpuTypeId := cpuType["id"].(string)
		cpuTypeName := cpuType["name"].(string)
		cpuTypeRamMB := cpuType["ramMB"].([]interface{})
		availableCpuTypes = append(availableCpuTypes, fmt.Sprintf("cpu_type=\"%s\" (%s)\ncpu_cores=%v\nram_mb=%v", cpuTypeId, cpuTypeName, cpuTypeCores, cpuTypeRamMB))
		if cpuTypeId == d.Get("cpu_type").(string) {
			validCores := false
			for _, cores := range cpuTypeCores {
				if cores.(float64) == d.Get("cpu_cores").(float64) {
					validCores = true
				}
			}
			if validCores {
				validRam := false
				for _, ramMb := range cpuTypeRamMB {
					if ramMb.(float64) == d.Get("ram_mb").(float64) {
						validRam = true
					}
				}
				if validRam {
					validCpuType = true
				}
			}
		}
	}
	var availableMonthlyTrafficPackages []string
	validMonthlyTrafficPackage := false
	for packageId, packageDescription := range monthlyTrafficPackage {
		availableMonthlyTrafficPackages = append(availableMonthlyTrafficPackages, fmt.Sprintf("monthly_traffic_package=\"%s\" (%s)", packageId, packageDescription.(string)))
		if packageId == d.Get("monthly_traffic_package").(string) {
			validMonthlyTrafficPackage = true
		}
	}
	if !validMonthlyTrafficPackage && d.Get("monthly_traffic_package").(string) == "" {
		if d.Get("billing_cycle").(string) == "monthly" {
			d.Set("monthly_traffic_package", defaultMonthlyTrafficPackage)
		}
		validMonthlyTrafficPackage = true
	}
	if d.Get("billing_cycle").(string) == "hourly" && d.Get("monthly_traffic_package") != "" {
		return errors.New("for hourly billing cycle, monthly traffic package must not be set")
	}

	disks := d.Get("disks").([]string)
	var disksFloat64 []float64
	for _, d := range disks {
		d = strings.TrimSuffix(d, "GB")
		d = strings.TrimSuffix(d, "gb")
		d = strings.TrimSuffix(d, "TB")
		d = strings.TrimSuffix(d, "tb")
		if val, err := strconv.ParseFloat(d, 64); err == nil {
			disksFloat64 = append(disksFloat64, val)
		}
	}
	validDisk := helper.Subslice(disksFloat64, diskSizeGB)

	if validCpuType && validMonthlyTrafficPackage && validDisk {
		id := []string{
			d.Get("datacenter_id").(string),
			d.Get("cpu_type").(string),
			fmt.Sprintf("%v", d.Get("cpu_cores").(float64)),
			fmt.Sprintf("%v", d.Get("ram_mb").(float64)),
			d.Get("billing_cycle").(string),
			d.Get("monthly_traffic_package").(string),
		}
		id = append(id, disks...)
		d.SetId(strings.Join(id, ","))
		return nil
	}

	d.SetId("")
	return errors.New(fmt.Sprintf("invalid server options, available options:\n\n%s\n\n%s\n\ndisks=[%s]",
		strings.Join(availableCpuTypes, "\n\n"),
		strings.Join(availableMonthlyTrafficPackages, "\n"),
		strings.Join(disks, ", "),
	))

}
