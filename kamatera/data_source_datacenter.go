package kamatera

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatacenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceDatacenterRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"country": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getDatacenterMatchesBy(datacenters map[string]map[string]string, attr string, value string) []string {
	var matchIDs []string
	if value != "" {
		for datacenterId, datacenter := range datacenters {
			if datacenter[attr] == value {
				matchIDs = append(matchIDs, datacenterId)
			}
		}
	}
	return matchIDs
}

func getAvailableDatacenters(datacenters map[string]map[string]string) string {
	var availableDatacenters []string
	availableDatacenters = append(availableDatacenters, fmt.Sprintf(
		"%-8s %-15s %-15s", "id", "country", "name",
	))
	for datacenterId, datacenter := range datacenters {
		availableDatacenters = append(availableDatacenters, fmt.Sprintf(
			"%-8s %-15s %-15s",
			"\""+datacenterId+"\"",
			"\""+datacenter["country"]+"\"",
			"\""+datacenter["name"]+"\"",
		))
	}
	return strings.Join(availableDatacenters, "\n")
}

func DataSourceDatacenterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*ProviderConfig)

	result, err := request(provider, "GET", "service/server?datacenter=1", nil)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	datacenters := map[string]map[string]string{}
	for _, datacenter := range result.([]interface{}) {
		datacenters[datacenter.(map[string]interface{})["id"].(string)] = map[string]string{
			"name":    datacenter.(map[string]interface{})["subCategory"].(string),
			"country": datacenter.(map[string]interface{})["name"].(string),
		}
	}

	id := d.Get("id").(string)
	country := d.Get("country").(string)
	name := d.Get("name").(string)

	datacenter, hasDatacenter := datacenters[id]

	countryDatacenterIDs := getDatacenterMatchesBy(datacenters, "country", country)
	nameDatacenterIDs := getDatacenterMatchesBy(datacenters, "name", name)
	if hasDatacenter &&
		(len(countryDatacenterIDs) == 0 || (len(countryDatacenterIDs) == 1 && countryDatacenterIDs[0] == datacenter["id"] && datacenter["country"] == country)) &&
		(len(nameDatacenterIDs) == 0 || (len(nameDatacenterIDs) == 1 && nameDatacenterIDs[0] == datacenter["id"] && datacenter["name"] == name)) {
		d.SetId(datacenter["id"])
		d.Set("name", datacenter["name"])
		d.Set("country", datacenter["country"])
		return nil
	} else if len(countryDatacenterIDs) == 1 &&
		(!hasDatacenter || datacenter["country"] == country) &&
		(len(nameDatacenterIDs) == 0 || (len(nameDatacenterIDs) == 1 && nameDatacenterIDs[0] == countryDatacenterIDs[0])) {

		d.SetId(countryDatacenterIDs[0])
		d.Set("name", datacenters[countryDatacenterIDs[0]]["name"])
		d.Set("country", country)
		return nil
	} else if len(nameDatacenterIDs) == 1 &&
		(!hasDatacenter || datacenter["name"] == name) {
		d.SetId(nameDatacenterIDs[0])
		d.Set("name", name)
		d.Set("country", datacenters[nameDatacenterIDs[0]]["country"])
		return nil
	} else {
		d.SetId("")
		d.Set("name", "")
		d.Set("country", "")
		return diag.Errorf("could not find matching datacenter, available datacenters: \n%s", getAvailableDatacenters(datacenters))
	}
}
