package kamatera

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The server name.",
			},
			"datacenter_id": {
				Type:     schema.TypeString,
				Required: true,
				Description: "id attribute of datacenter data source.",
			},
			"cpu_type": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The CPU type - a single upper-case letter. See https://console.kamatera.com/pricing for " +
					"available CPU types and description of each type.",
			},
			"cpu_cores": {
				Type:     schema.TypeFloat,
				Optional: true,
				Description: "Number of CPU cores to allocate. See https://console.kamatera.com/pricing for a " +
					"a description of the meaning of this value depending on the selected CPU type.",
			},
			"ram_mb": {
				Type:     schema.TypeFloat,
				Optional: true,
				Description: "Amount of RAM to allocate in MB.",
			},
			"disk_sizes_gb": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeFloat},
				MinItems: 1,
				MaxItems: 4,
				Optional: true,
				Description: "List of disk sizes in GB, each item in the list will create a new disk in given " +
					"size and attach it to the server.",
			},
			"billing_cycle": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "hourly",
				Description: "hourly or monthly, see https://console.kamatera.com/pricing for details.",
			},
			"monthly_traffic_package": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "For advanced use-cases you can select a specific traffic package, depending on " +
					"datacenter availability. See https://console.kamatera.com/pricing for details.",
			},
			"power_on": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "true by default, set to false to have the server created without powering it on.",
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				Description: "id attribute of image data source",
			},
			"network": {
				Type:     schema.TypeList,
				MaxItems: 4,
				Description: "Network interfaces to attach to the server. If not specified a single WAN interface with " +
					"auto IP will be attached.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Set to 'wan' to attach a public internet interface with auto-allocated IP. " +
								"To use a private network, set to full_name attribute of network data source",
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "auto",
							Description: "The IP to use, leave unset or set to 'auto' to auto-allocate an IP",
						},
					},
				},
				Optional: true,
			},
			"daily_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Set to true to enable daily backups.",
			},
			"managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Set to true for managed support services.",
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: "The server root password.",
			},
			"ssh_pubkey": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "SSH public key to allow access to the server without a password.",
			},
			"generated_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				Description: "In case password was not provided, an auto-generated password will be used.",
			},
			"price_monthly_on": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The monthly price if server is turned on for the entire month.",
			},
			"price_hourly_on": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The hourly price if server is turned on for the entire hour.",
			},
			"price_hourly_off": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The hourly price if server is turned off for the entire hour.",
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
			"startup_script": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
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
			diskSizesGB = append(diskSizesGB, fmt.Sprintf("size=%v", v))
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
		RAM:              d.Get("ram_mb").(float64),
		Disk:             strings.Join(diskSizesGB, " "),
		DailyBackup:      dailyBackup,
		Managed:          managed,
		Network:          strings.Join(networks, " "),
		Quantity:         "1",
		BillingCycle:     d.Get("billing_cycle").(string),
		MonthlyPackage:   d.Get("monthly_traffic_package").(string),
		PowerOn:          powerOn,
		ScriptFile:       d.Get("startup_script").(string),
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
		return diag.FromErr(err)
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
	return resourceServerRead(ctx, d, m)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
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
	{
		cpuCores, err := strconv.ParseFloat(cpu[0:1], 16)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("cpu_cores", cpuCores)
	}

	{
		diskSizes := server["diskSizes"].([]interface{})
		var diskSizesString []float64
		for _, v := range diskSizes {
			diskSizesString = append(diskSizesString, v.(float64))
		}
		d.Set("disk_sizes_gb", diskSizesString)
	}

	d.Set("power_on", server["power"].(string) == "on")
	d.Set("datacenter_id", server["datacenter"].(string))
	d.Set("ram_mb", server["ram"].(float64))
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

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	newCPU := ""
	{
		var newCPUType interface{}
		{
			o, n := d.GetChange("cpu_type")
			if d.HasChange("cpu_type") {
				newCPUType = n
			} else {
				newCPUType = o
			}
		}
		var newCPUCores interface{}
		{
			o, n := d.GetChange("cpu_cores")
			if d.HasChange("cpu_cores") {
				newCPUCores = n
			} else {
				newCPUCores = o
			}
		}
		if d.HasChanges("cpu_type", "cpu_cores") {
			newCPU = fmt.Sprintf("%v%v", newCPUCores, newCPUType)
		}
	}

	var newRAM float64
	if d.HasChange("ram_mb") {
		_, n := d.GetChange("ram_mb")
		newRAM = n.(float64)
	}

	if d.HasChange("image_id") {
		// TODO: Implement
		return diag.Errorf("changing server image is not supported yet")
	}

	if d.HasChange("network") {
		// TODO: Implement
		return diag.Errorf("changing server networks is not supported yet")
	}

	if d.HasChange("ssh_pubkey") {
		// TODO: implement
		return diag.Errorf("changing server ssh_pubkey is not supported yet")
	}

	if d.HasChange("startup_script") {
		return diag.Errorf("changing server startup_script is not supported")
	}

	oldBillingCycle := ""
	newBillingCycle := ""
	if d.HasChange("billing_cycle") {
		o, n := d.GetChange("billing_cycle")
		oldBillingCycle = o.(string)
		newBillingCycle = n.(string)
	}

	oldTrafficPackage := ""
	newTrafficPackage := ""
	if d.HasChange("monthly_traffic_package") {
		o, n := d.GetChange("monthly_traffic_package")
		oldTrafficPackage = o.(string)
		newTrafficPackage = n.(string)
	}

	if d.HasChange("datacenter_id") {
		return diag.Errorf("changing datacenter is not supported yet")
	}

	newDailyBackup := ""
	if d.HasChange("daily_backup") {
		newDailyBackup = "no"
		if d.Get("daily_backup").(bool) {
			newDailyBackup = "yes"
		}
	}

	newManaged := ""
	if d.HasChange("managed") {
		newManaged = "no"
		if d.Get("managed").(bool) {
			newManaged = "yes"
		}
	}

	provider := m.(*ProviderConfig)
	if err := serverConfigure(
		provider,
		d.Get("internal_server_id").(string),
		newCPU,
		newRAM,
		oldTrafficPackage, newTrafficPackage,
		oldBillingCycle, newBillingCycle,
		newDailyBackup,
		newManaged,
	); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("disk_sizes_gb") {
		o, n := d.GetChange("disk_sizes_gb")

		op, err := calDiskChangeOperation(o, n)
		if err != nil {
			return diag.FromErr(err)
		}

		err = changeDisks(provider, d.Get("internal_server_id").(string), op)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		o, n := d.GetChange("password")

		err := serverChangePassword(provider, d.Get("internal_server_id").(string), n.(string))
		if err != nil {
			d.Set("password", o)
			return diag.FromErr(err)
		}

		d.Set("password", n)
	}

	if d.HasChange("name") {
		_, n := d.GetChange("name")
		if err := renameServer(provider, d.Get("internal_server_id").(string), n.(string)); err != nil {
			return diag.FromErr(err)
		}
		d.Set("name", n)
	}

	if d.HasChange("power_on") {
		if d.Get("power_on").(bool) {
			if err := changeServerPower(provider, d.Get("internal_server_id").(string), "poweron"); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := changeServerPower(provider, d.Get("internal_server_id").(string), "poweroff"); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*ProviderConfig)
	err := changeServerPower(provider, d.Get("internal_server_id").(string), "terminate")
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServerImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	d.Set("internal_server_id", d.Id())
	diags := resourceServerRead(ctx, d, m)
	if diags.HasError() {
		var errorMessages []string
		for i := range diags {
			errorMessages = append(errorMessages, diags[i].Summary)
		}
		return nil, fmt.Errorf(strings.Join(errorMessages, ", "))
	}
	return []*schema.ResourceData{d}, nil
}

func serverConfigure(
	provider *ProviderConfig, internalServerId string, newCpu string, newRam float64,
	oldTrafficPackage string, newTrafficPackage string, oldBillingCycle string, newBillingCycle string,
	newDailyBackup string, newManaged string,
) error {
	if newCpu != "" {
		if e := postServerConfigure(
			provider,
			configureServerPostValues{ID: internalServerId, CPU: newCpu},
		); e != nil {
			return e
		}
	}

	if newRam != 0 {
		if e := postServerConfigure(
			provider,
			configureServerPostValues{ID: internalServerId, RAM: newRam},
		); e != nil {
			return e
		}
	}

	if newTrafficPackage != "" || newBillingCycle != "" {
		billingCycle := ""
		if newBillingCycle != "" {
			billingCycle = newBillingCycle
		}
		trafficPackage := oldTrafficPackage
		if newTrafficPackage != "" {
			trafficPackage = newTrafficPackage
		}
		if e := postServerConfigure(
			provider,
			configureServerPostValues{ID: internalServerId, MonthlyPackage: trafficPackage, BillingCycle: billingCycle},
		); e != nil {
			return e
		}
	}

	if newDailyBackup != "" {
		if e := postServerConfigure(
			provider,
			configureServerPostValues{ID: internalServerId, DailyBackup: newDailyBackup},
		); e != nil {
			return e
		}
	}

	if newManaged != "" {
		if e := postServerConfigure(
			provider,
			configureServerPostValues{ID: internalServerId, Managed: newManaged},
		); e != nil {
			return e
		}
	}

	return nil
}

func changeServerPower(provider *ProviderConfig, internalServerID string, operation string) error {
	var body powerOperationServerPostValues
	if operation == "terminate" {
		body = powerOperationServerPostValues{ID: internalServerID, Force: true}
	} else {
		body = powerOperationServerPostValues{ID: internalServerID}
	}

	result, err := request(provider, "POST", fmt.Sprintf("service/server/%s", operation), body)
	if err != nil {
		return err
	}

	commandIds := result.([]interface{})
	_, err = waitCommand(provider, commandIds[0].(string))

	return err
}