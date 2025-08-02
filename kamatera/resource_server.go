package kamatera

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		CustomizeDiff: resourceServerCustomizeDiff,
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServerImport,
		},
		Description: "It's recommended to use our " +
			"[server configuration interface]" +
			"(https://kamatera.github.io/kamateratoolbox/serverconfiggen.html?configformat=terraform) " +
			"which provides ready to use Terraform templates with valid configuration options and identifiers " +
			"according to your selection.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The server name.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(4, 40),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-.]*$`), "must contain only letters, digits, dashes (-) and dots (.)"),
				),
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "id attribute of datacenter data source.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 6),
					validation.StringMatch(regexp.MustCompile(`^[A-Z0-9-]+$`), "must contain only uppercase letters, digits and dashes (-)"),
				),
			},
			"cpu_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "B",
				Description: "The CPU type - a single upper-case letter. See https://console.kamatera.com/pricing for " +
					"available CPU types and description of each type.",
				ValidateFunc: validation.All(
					validation.StringInSlice([]string{"A", "B", "T", "D"}, false),
				),
			},
			"cpu_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
				Description: "Number of CPU cores to allocate. See https://console.kamatera.com/pricing for a " +
					"a description of the meaning of this value depending on the selected CPU type.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ram_mb": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1024,
				Description:  "Amount of RAM to allocate in MB.",
				ValidateFunc: validation.IntAtLeast(256),
			},
			"disk_sizes_gb": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				MinItems: 1,
				MaxItems: 4,
				Optional: true,
				Description: "List of disk sizes in GB, each item in the list will create a new disk in given " +
					"size and attach it to the server.",
				DefaultFunc: func() (interface{}, error) {
					return []interface{}{10}, nil
				},
			},
			"billing_cycle": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "hourly",
				Description:  "hourly or monthly, see https://console.kamatera.com/pricing for details.",
				ValidateFunc: validation.StringInSlice([]string{"hourly", "monthly"}, false),
			},
			"monthly_traffic_package": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "For advanced use-cases you can select a specific traffic package, depending on " +
					"datacenter availability. See https://console.kamatera.com/pricing for details.",
			},
			"power_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "true by default, set to false to have the server created without powering it on.",
			},
			"image_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "id attribute of image data source",
			},
			"network": {
				Type:     schema.TypeList,
				MaxItems: 4,
				ForceNew: true,
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
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "auto",
							Description: "The IP to use, leave unset or set to 'auto' to auto-allocate an IP",
						},
					},
				},
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return []interface{}{
						map[string]interface{}{
							"name": "wan",
						},
					}, nil
				},
			},
			"daily_backup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to true to enable daily backups.",
			},
			"managed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to true for managed support services.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The server root password.",
			},
			"ssh_pubkey": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "SSH public key to allow access to the server without a password.",
			},
			"generated_password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "In case password was not provided, an auto-generated password will be used.",
			},
			"price_monthly_on": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The monthly price if server is turned on for the entire month.",
			},
			"price_hourly_on": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The hourly price if server is turned on for the entire hour.",
			},
			"price_hourly_off": {
				Type:        schema.TypeString,
				Computed:    true,
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
				ForceNew: true,
			},
			"allow_recreate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to true to allow recreation of the server for changes that require recreation. ",
			},
		},
	}
}

func resourceServerCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	err := loadServerOptions()
	if err != nil {
		return fmt.Errorf("failed to load server options: %w", err)
	}
	var errors []error
	err = serverOptionsValidateDatacenter(d.Get("datacenter_id").(string))
	if err != nil {
		errors = append(errors, err)
	}
	err = serverOptionsValidateCpu(fmt.Sprintf("%v%v", d.Get("cpu_cores"), d.Get("cpu_type")))
	if err != nil {
		errors = append(errors, err)
	}
	err = serverOptionsValidateRamMB(d.Get("cpu_type").(string), d.Get("ram_mb").(int))
	if err != nil {
		errors = append(errors, err)
	}
	for _, diskSize := range d.Get("disk_sizes_gb").([]interface{}) {
		err = serverOptionsValidateDiskSizeGB(diskSize.(int))
		if err != nil {
			errors = append(errors, err)
		}
	}
	if d.Get("billing_cycle").(string) == "monthly" {
		err = serverOptionsValidateMonthlyTrafficPackage(d.Get("datacenter_id").(string), d.Get("monthly_traffic_package").(string))
		if err != nil {
			errors = append(errors, err)
		}
	} else if d.Get("billing_cycle").(string) != "hourly" {
		errors = append(errors, fmt.Errorf("billing cycle must be either 'hourly' or 'monthly', got '%s'", d.Get("billing_cycle").(string)))
	}
	if len(errors) > 0 {
		var errorMessages []string
		for _, e := range errors {
			errorMessages = append(errorMessages, e.Error())
		}
		return fmt.Errorf("invalid server configuration: %s", strings.Join(errorMessages, ", "))
	} else {
		return nil
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
		RAM:              d.Get("ram_mb").(int),
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
	d.Set("cpu_type", cpu[len(cpu)-1:])
	{
		cpuCores, err := strconv.ParseInt(cpu[:len(cpu)-1], 16, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("cpu_cores", int(cpuCores))
	}

	{
		diskSizes := server["diskSizes"].([]interface{})
		var diskSizesString []int
		for _, v := range diskSizes {
			var intv int
			switch v.(type) {
			case int:
				intv = v.(int)
			case float64:
				intv = int(v.(float64))
			}
			diskSizesString = append(diskSizesString, intv)
		}
		d.Set("disk_sizes_gb", diskSizesString)
	}

	d.Set("power_on", server["power"].(string) == "on")
	d.Set("datacenter_id", server["datacenter"].(string))

	var intram int
	switch server["ram"].(type) {
	case int:
		intram = server["ram"].(int)
	case float64:
		intram = int(server["ram"].(float64))
	}
	d.Set("ram_mb", intram)

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

	var newRAM int
	if d.HasChange("ram_mb") {
		_, n := d.GetChange("ram_mb")
		newRAM = n.(int)
	}

	if d.HasChange("image_id") && !d.Get("allow_recreate").(bool) {
		return diag.Errorf("changing server image requires recreation, set allow_recreate to true to allow this change")
	}

	if d.HasChange("network") && !d.Get("allow_recreate").(bool) {
		return diag.Errorf("changing server networks requires recreation, set allow_recreate to true to allow this change")
	}

	if d.HasChange("ssh_pubkey") && !d.Get("allow_recreate").(bool) {
		return diag.Errorf("changing server ssh_pubkey requires recreation, set allow_recreate to true to allow this change")
	}

	if d.HasChange("startup_script") && !d.Get("allow_recreate").(bool) {
		return diag.Errorf("changing server startup_script requires recreation, set allow_recreate to true to allow this change")
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

	if d.HasChange("datacenter_id") && !d.Get("allow_recreate").(bool) {
		return diag.Errorf("changing datacenter requires recreation, set allow_recreate to true to allow this change")
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
	provider *ProviderConfig, internalServerId string, newCpu string, newRam int,
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
