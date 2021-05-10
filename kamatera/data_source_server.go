package kamatera

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: dataSourceServerCreate,
		ReadContext: dataSourceServerRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"datacenter_id": &schema.Schema{
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
			"ram_mb": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"disk_sizes_gb": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeFloat},
				MinItems: 1,
				MaxItems: 4,
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

func dataSourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)

	password := d.Get("password").(string)
	if password == "" {
		password = "__generate__"
	}

	dailyBackup := "no"
	if d.Get("daily_backup").(bool) {
		dailyBackup = "yes"
	}

	managed := "no"
	if d.Get("managed").(bool) {
		managed = "yes"
	}

	var networks []string
	for _, network := range d.Get("network").([]interface{}) {
		network := network.(map[string]interface{})
		networks = append(networks, fmt.Sprintf("name=%s,ip=%s", network["name"].(string), network["ip"].(string)))
	}
	if len(networks) == 0 {
		networks = append(networks, "name=wan,ip=auto")
	}

	powerOn := "no"
	if d.Get("power_on").(bool) {
		powerOn = "yes"
	}

	var diskSizesGB []string
	{
		diskSizes := d.Get("disk_sizes_gb").([]interface{})
		for _, v := range diskSizes {
			diskSizesGB = append(diskSizesGB, v.(string))
		}
	}

	body := &createServerPostValues{
		Name:             d.Get("name").(string),
		Password:         password,
		PasswordValidate: password,
		SSHKey:           d.Get("ssh_pubkey").(string),
		Datacenter:       d.Get("datacenter_id").(string),
		Image:            d.Get("image_id").(string),
		CPU:              fmt.Sprintf("%v%v", d.Get("cpu_cores"), d.Get("cpu_type")),
		RAM:              d.Get("ram_mb").(int64),
		Disk:             strings.Join(diskSizesGB, " "),
		DailyBackup:      dailyBackup,
		Managed:          managed,
		Network:          strings.Join(networks, " "),
		Quantity:         "1",
		BillingCycle:     d.Get("billing_cycle").(string),
		MonthlyPackage:   d.Get("monthly_traffic_package").(string),
		PowerOn:          powerOn,
	}
	result, err := request(provider, "POST", "service/server", body)
	if err != nil {
		return diag.FromErr(err)
	}

	var commandIDs []interface{}
	if password == "__generate__" {
		response := result.(map[string]interface{})
		d.Set("generated_password", response["password"].(string))
		commandIDs = response["commandIds"].([]interface{})
	} else {
		d.Set("generated_password", "")
		commandIDs = result.([]interface{})
	}

	if len(commandIDs) != 1 {
		return diag.Errorf("invalid response from Kamatera API: did not return expected command ID")
	}

	commandID := commandIDs[0].(string)
	command, err := waitCommand(provider, commandID)
	if err != nil {
		diag.FromErr(err)
	}

	createLog, hasCreateLog := command["log"]
	if !hasCreateLog {
		return diag.Errorf("invalid response from Kamatera API: command is missing creation log")
	}

	createdServerName := ""
	for _, line := range strings.Split(createLog.(string), "\n") {
		if strings.HasPrefix(line, "Name: ") {
			createdServerName = strings.Replace(line, "Name: ", "", 1)
		}
	}
	if createdServerName == "" {
		return diag.Errorf("invalid response from Kamatera API: failed to get created server name")
	}
	d.SetId(createdServerName)
	return dataSourceServerCreate(ctx, d, m)
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)
	var body listServersPostValues

	if d.Get("internal_server_id").(string) == "" {
		body = listServersPostValues{Name: d.Id()}
	} else {
		body = listServersPostValues{ID: d.Get("internal_server_id").(string)}
	}
	result, err := request(provider, "POST", fmt.Sprintf("service/server/info"), body)
	if err != nil {
		return diag.FromErr(err)
	}

	servers := result.([]interface{})
	if len(servers) != 1 {
		return diag.Errorf("failed to find server")
	}
	server := servers[0].(map[string]interface{})

	d.Set("id", server["id"].(string))

	d.Set("name", server["name"].(string))

	cpu := server["cpu"].(string)
	d.Set("cpu_type", cpu[1:2])
	d.Set("cpu_cores", cpu[0:1])

	{
		diskSizes := server["diskSizes"].([]interface{})
		var diskSizesString []string
		for _, v := range diskSizes {
			diskSizesString = append(diskSizesString, v.(string))
		}
		d.Set("disk_sizes_gb", diskSizesString)
	}

	d.Set("power_on", server["power"].(string) == "on")
	d.Set("datacenter_id", server["datacenter"].(string))
	d.Set("ram_mb", server["ram"].(int64))
	d.Set("daily_backup", server["backup"].(string) == "1")
	d.Set("managed", server["managed"].(string) == "1")
	d.Set("billing_cycle", server["billing"].(string))
	d.Set("monthly_traffic_package", server["traffic"].(string))
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
