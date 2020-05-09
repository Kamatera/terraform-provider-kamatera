package main

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func dataSourceDatacenter() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceDatacenterRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"country": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getDatacenterMatchesBy(datacenters map[string]map[string]string, attr string, value string) []string {
	var matchIds []string
	if value != "" {
		for datacenterId, datacenter := range datacenters {
			if datacenter[attr] == value {
				matchIds = append(matchIds, datacenterId)
			}
		}
	}
	return matchIds
}

func getAvailableDatacenters(datacenters map[string]map[string]string) string {
	var availableDatacenters []string
	availableDatacenters = append(availableDatacenters, fmt.Sprintf(
		"%-8s %-15s %-15s", "id", "country", "name",
	))
	for datacenterId, datacenter := range datacenters {
		availableDatacenters = append(availableDatacenters, fmt.Sprintf(
			"%-8s %-15s %-15s",
			"\"" + datacenterId + "\"",
			"\"" + datacenter["country"] + "\"",
			"\"" + datacenter["name"] + "\"",
		))
	}
	return strings.Join(availableDatacenters, "\n")
}

func DataSourceDatacenterRead(d *schema.ResourceData, m interface{}) error {
	provider := m.(*ProviderConfiguration)
	result, e := kamateraRequest(*provider, "GET", "service/server?datacenter=1", nil)
	if e != nil {
		d.SetId("")
		return e
	}
	datacenters := map[string]map[string]string{}
	for _, datacenter := range result.([]interface{}) {
		datacenters[datacenter.(map[string]interface{})["id"].(string)] = map[string]string{
			"name": datacenter.(map[string]interface{})["subCategory"].(string),
			"country": datacenter.(map[string]interface{})["name"].(string),
		}
	}
	id := d.Get("id").(string)
	country := d.Get("country").(string)
	name := d.Get("name").(string)
	datacenter, hasDatacenter := datacenters[id]
	countryDatacenterIds := getDatacenterMatchesBy(datacenters, "country", country)
	nameDatacenterIds := getDatacenterMatchesBy(datacenters, "name", name)
	if hasDatacenter &&
			(len(countryDatacenterIds) == 0 || (len(countryDatacenterIds) == 1 && countryDatacenterIds[0] == datacenter["id"] && datacenter["country"] == country)) &&
			(len(nameDatacenterIds) == 0    || (len(nameDatacenterIds)    == 1 && nameDatacenterIds[0]    == datacenter["id"] && datacenter["name"]    == name   )) {
		d.SetId(datacenter["id"])
		d.Set("name", datacenter["name"])
		d.Set("country", datacenter["country"])
		return nil
	} else if len(countryDatacenterIds) == 1 &&
			(! hasDatacenter || datacenter["country"] == country) &&
			(len(nameDatacenterIds) == 0    || (len(nameDatacenterIds)    == 1 && nameDatacenterIds[0]    == countryDatacenterIds[0])) {

		d.SetId(countryDatacenterIds[0])
		d.Set("name", datacenters[countryDatacenterIds[0]]["name"])
		d.Set("country", country)
		return nil
	} else if len(nameDatacenterIds) == 1 &&
			(! hasDatacenter || datacenter["name"] == name) {
		d.SetId(nameDatacenterIds[0])
		d.Set("name", name)
		d.Set("country", datacenters[nameDatacenterIds[0]]["country"])
		return nil
	} else {
		d.SetId("")
		d.Set("name", "")
		d.Set("country", "")
		return errors.New(fmt.Sprintf("could not find matching datacenter, available datacenters: \n%s", getAvailableDatacenters(datacenters)))
	}
}
