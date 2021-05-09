package kamatera

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kamatera/terraform-provider-kamatera/kamatera/helper"
)

func dataSourceServer() *schema.Resource {
	return &schema.Resource{
		// ReadContext: dataSourceCoffeesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cpu_cores": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"power_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network": {
				Type:     schema.TypeList,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "auto",
						},
					},
				},
				Optional: true,
			},
			"daily_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"ssh_pubkey": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"generated_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"price_monthly_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"price_hourly_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"price_hourly_off": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_networks": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ips": {
							Type:     schema.TypeList,
							Elem:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"public_ips": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"private_ips": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"internal_server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)
	var body listServersPostValues

	if d.Get("internal_server_id").(string) == "" {
		body = listServersPostValues{Name: d.Id()}
	} else {
		body = listServersPostValues{ID: d.Get("internal_server_id").(string)}
	}
	result, err := helper.Request(provider, "POST", fmt.Sprintf("service/server/info"), body)
	if err != nil {
		return diag.FromErr(err)
	}

	servers := result.([]interface{})
	if len(servers) != 1 {
		return diag.Errorf("failed to find server")
	}
	server := servers[0].(map[string]interface{})

	d.Set("id", server["id"].(string))

	name := server["name"].(string)
	d.Set("name", name)

	cpu := server["cpu"].(string)
	d.Set("cpu_type", cpu[1:2])
	d.Set("cpu_cores", cpu[0:1])
	diskSizes := server["diskSizes"].([]interface{})

	power := server["power"].(string)
	d.Set("power_on", power == "on")

	datacenter := server["datacenter"].(string)
	ram := server["ram"].(float64)
	backup := server["backup"].(string)
	managed := server["managed"].(string)
	billing := server["billing"].(string)
	traffic := server["traffic"].(string)

	d.Set("managed", managed == "1")
	d.Set("daily_backup", backup == "1")
	d.Set("internal_server_id", server["id"].(string))
	d.Set("price_monthly_on", server["priceMonthlyOn"].(string))
	d.Set("price_hourly_on", server["priceHourlyOn"].(string))
	d.Set("price_hourly_off", server["priceHourlyOff"].(string))

	networks := server["networks"].([]interface{})
	var publicIPs []string
	var privateIPs []string
	var attachedNetworks []interface{}
	for _, network := range networks {
		network := network.(map[string]interface{})
		attachedNetworks = append(attachedNetworks, network)
		if strings.Index(network["network"].(string), "wan-") == 0 {
			for _, ip := range network["ips"].([]interface{}) {
				publicIPs = append(publicIPs, ip.(string))
			}
		} else {
			for _, ip := range network["ips"].([]interface{}) {
				privateIPs = append(privateIPs, ip.(string))
			}
		}
	}
	d.Set("public_ips", publicIPs)
	d.Set("private_ips", privateIPs)
	d.Set("attached_networks", attachedNetworks)

	return
}
