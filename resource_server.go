package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"power_on": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: true,
			},
			"server_options_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
			},
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"network": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 4,
				Elem:     &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required: true,
						},
						"ip": &schema.Schema{
							Type:        schema.TypeString,
							Optional: true,
							Default: "auto",
						},
					},
				},
				Optional: true,
			},
			"daily_backup": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"managed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Sensitive: true,
			},
			"ssh_pubkey": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
			},
			"generated_password": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
				Sensitive: true,
			},
			"internal_server_id": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			"price_monthly_on": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			"price_hourly_on": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			"price_hourly_off": &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			"attached_networks": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": &schema.Schema{
							Type: schema.TypeString,
							Computed: true,
						},
						"ips": &schema.Schema{
							Type: schema.TypeList,
							Elem: schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"public_ips": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString,},
				Computed: true,
			},
			"private_ips": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString,},
				Computed: true,
			},
		},
	}
}

type CreateServerPostValues struct {
	Name             string `json:"name"`
	Password         string `json:"password"`
	PasswordValidate string `json:"passwordValidate"`
	SshKey           string `json:"ssh-key"`
	Datacenter       string `json:"datacenter"`
	Image            string `json:"image"`
	Cpu              string `json:"cpu"`
	Ram              int64  `json:"ram"`
	Disk             string `json:"disk"`
	DailyBackup      string `json:"dailybackup"`
	Managed          string `json:"managed"`
	Network          string `json:"network"`
	Quantity         string `json:"quantity"`
	BillingCycle     string `json:"billingcycle"`
	MonthlyPackage   string `json:"monthlypackage"`
	PowerOn          string `json:"poweronaftercreate"`
}

func parseServerOptions(serverOptionsId string) (map[string]interface{}, error) {
	res :=  make(map[string]interface{})
	serverOptions := strings.Split(serverOptionsId, ",")
	res["datacenter"] = serverOptions[0]
	res["cpuType"] = serverOptions[1]
	res["cpuCores"] = serverOptions[2]
	res["ramMB"], _ = strconv.ParseInt(serverOptions[3], 10, 0)
	res["billingCycle"] = serverOptions[4]
	res["monthlyTrafficPackage"] = serverOptions[5]
	var diskSizesGB []string
	for _, sizeGB := range serverOptions[6:] {
		diskSizesGB = append(diskSizesGB, fmt.Sprintf("size=%s", sizeGB))
	}
	res["diskSizesGB"] = diskSizesGB
	return res, nil
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	serverOptions, e := parseServerOptions(d.Get("server_options_id").(string))
	if e != nil {
		return e
	}
	datacenter := serverOptions["datacenter"].(string)
	cpuType := serverOptions["cpuType"].(string)
	cpuCores := serverOptions["cpuCores"].(string)
	ramMB := serverOptions["ramMB"].(int64)
	billingCycle := serverOptions["billingCycle"].(string)
	monthlyTrafficPackage := serverOptions["monthlyTrafficPackage"].(string)
	diskSizesGB := serverOptions["diskSizesGB"].([]string)
	password := d.Get("password").(string)
	if password == "" {
		password = "__generate__"
	}
	ssh_pubkey := d.Get("ssh_pubkey").(string)
	provider := m.(*ProviderConfiguration)
	daily_backup := "no"
	if d.Get("daily_backup").(bool) {
		daily_backup = "yes"
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
	body := &CreateServerPostValues{
		Name:             d.Get("name").(string),
		Password:         password,
		PasswordValidate: password,
		SshKey:           ssh_pubkey,
		Datacenter:       datacenter,
		Image:            d.Get("image_id").(string),
		Cpu:              fmt.Sprintf("%s%s", cpuCores, cpuType),
		Ram:              ramMB,
		Disk:             strings.Join(diskSizesGB, " "),
		DailyBackup:      daily_backup,
		Managed:          managed,
		Network:          strings.Join(networks, " "),
		Quantity:         "1",
		BillingCycle:     billingCycle,
		MonthlyPackage:   monthlyTrafficPackage,
		PowerOn:          powerOn,
	}
	result, e := kamateraRequest(*provider, "POST", "service/server", body)
	if e != nil {
		return e
	}
	var commandIds []interface{}
	if password == "__generate__" {
		response := result.(map[string]interface{})
		d.Set("generated_password", response["password"].(string))
		commandIds = response["commandIds"].([]interface{})
	} else {
		d.Set("generated_password", "")
		commandIds = result.([]interface{})
	}
	if len(commandIds) != 1 {
		return errors.New("invalid response from Kamatera API: did not return expected command ID")
	}
	commandId := commandIds[0].(string)
	command, e := waitCommand(*provider, commandId)
	if e != nil {
		return e
	}
	createLog, hasCreateLog := command["log"]
	if ! hasCreateLog {
		return errors.New("invalid response from Kamatera API: command is missing creation log")
	}
	createdServerName := ""
	for _, line := range strings.Split(createLog.(string), "\n") {
		if strings.HasPrefix(line, "Name: ") {
			createdServerName = strings.Replace(line, "Name: ", "", 1)
		}
	}
	if createdServerName == "" {
		return errors.New("invalid response from Kamatera API: failed to get created server name")
	}
	d.SetId(createdServerName)
	return resourceServerRead(d, m)
}

type ListServersPostValues struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	provider := m.(*ProviderConfiguration)
	var body ListServersPostValues
	if d.Get("internal_server_id").(string) == "" {
		body = ListServersPostValues{Name:d.Id()}
	} else {
		body = ListServersPostValues{Id:d.Get("internal_server_id").(string)}
	}
	result, e := kamateraRequest(*provider, "POST", fmt.Sprintf("service/server/info"), body)
	if e != nil {
		return e
	}
	servers := result.([]interface{})
	if len(servers) != 1 {
		return errors.New("failed to find server")
	}
	server := servers[0].(map[string]interface{})
	cpu := server["cpu"].(string)
	diskSizes := server["diskSizes"].([]interface{})
	name := server["name"].(string)
	power := server["power"].(string)
	datacenter := server["datacenter"].(string)
	ram := server["ram"].(float64)
	networks := server["networks"].([]interface{})
	backup := server["backup"].(string)
	managed := server["managed"].(string)
	billing := server["billing"].(string)
	traffic := server["traffic"].(string)
	serverOptionsId := []string{datacenter, cpu[1:2], cpu[0:1], fmt.Sprintf("%v", ram), billing, traffic,}
	for _, diskSize := range diskSizes {
		serverOptionsId = append(serverOptionsId, fmt.Sprintf("%v", diskSize))
	}
	d.Set("server_options_id", strings.Join(serverOptionsId, ","))
	_managed := false
	if managed == "1" {
		_managed = true
	}
	d.Set("managed", _managed)
	_backup := false
	if backup == "1" {
		_backup = true
	}
	d.Set("daily_backup", _backup)
	d.SetId(name)
	d.Set("internal_server_id", server["id"].(string))
	d.Set("price_monthly_on", server["priceMonthlyOn"].(string))
	d.Set("price_hourly_on", server["priceHourlyOn"].(string))
	d.Set("price_hourly_off", server["priceHourlyOff"].(string))
	if power == "on" {
		d.Set("power_on", true)
	} else {
		d.Set("power_on", false)
	}
	var publicIps []string
	var privateIps []string
	var setNetworks []interface{}
	for _, network := range networks {
		network := network.(map[string]interface{})
		setNetworks = append(setNetworks, network)
		if strings.Index(network["network"].(string), "wan-") == 0 {
			for _, ip := range network["ips"].([]interface{}) {
				publicIps = append(publicIps, ip.(string))
			}
		} else {
			for _, ip := range network["ips"].([]interface{}) {
				privateIps = append(privateIps, ip.(string))
			}
		}
	}
	d.Set("public_ips", publicIps)
	d.Set("private_ips", privateIps)
	d.Set("attached_networks", setNetworks)
	return nil
}

type RenameServerPostValues struct {
	Id string `json:"id"`
	NewName string `json:"new-name"`
}

func serverRename(provider ProviderConfiguration, internalServerId string, name string) error {
	result, e := kamateraRequest(provider, "POST", fmt.Sprintf("service/server/rename"), &RenameServerPostValues{Id: internalServerId, NewName: name,})
	if e != nil {
		return e
	}
	commandIds := result.([]interface{})
	if _, e := waitCommand(provider, commandIds[0].(string)); e != nil {
		return e
	}
	return nil
}

type ChangePasswordServerPostValues struct {
	Id string `json:"id"`
	Password string `json:"password"`
}

func serverChangePassword(provider ProviderConfiguration, internalServerId string, password string) error {
	result, e := kamateraRequest(provider, "POST", "service/server/password", &ChangePasswordServerPostValues{Id: internalServerId, Password: password})
	if e != nil {
		return e
	}
	commandIds := result.([]interface{})
	if _, e := waitCommand(provider, commandIds[0].(string)); e != nil {
		return e
	}
	return nil
}

type PowerOperationServerPostValues struct {
	Id string `json:"id"`
	Force bool `json:"force"`
}

func serverPowerOperation(provider ProviderConfiguration, internalServerId string, operation string) error {
	var body PowerOperationServerPostValues
	if operation == "terminate" {
		body = PowerOperationServerPostValues{Id: internalServerId, Force: true}
	} else {
		body = PowerOperationServerPostValues{Id: internalServerId}
	}
	result, e := kamateraRequest(provider, "POST", fmt.Sprintf("service/server/%s", operation), body)
	if e != nil {
		return e
	}
	commandIds := result.([]interface{})
	if _, e := waitCommand(provider, commandIds[0].(string)); e != nil {
		return e
	}
	return nil
}

type ConfigureServerPostValues struct {
	Id string `json:"id"`
	Cpu string `json:"cpu"`
	Ram string `json:"ram"`
	DailyBackup string `json:"dailybackup"`
	Managed string `json:"managed"`
	BillingCycle string `json:"billingcycle"`
	MonthlyPackage string `json:"monthlypackage"`
}

func postServerConfigure(provider ProviderConfiguration, postValues ConfigureServerPostValues) error {
	result, e := kamateraRequest(provider, "POST", "server/configure", postValues)
	if e != nil {
		return e
	}
	commandIds := result.([]interface{})
	if _, e := waitCommand(provider, commandIds[0].(string)); e != nil {
		return e
	}
	return nil
}

func serverConfigure(
	provider ProviderConfiguration, internalServerId string, newCpu string, newRam string,
	oldTrafficPackage string, newTrafficPackage string, oldBillingCycle string, newBillingCycle string,
	newDailyBackup string, newManaged string,
	) error {
	if newCpu != "" {
		if e := postServerConfigure(provider, ConfigureServerPostValues{Id: internalServerId, Cpu: newCpu,}); e != nil {
			return e
		}
	}
	if newRam != "" {
		if e := postServerConfigure(provider, ConfigureServerPostValues{Id: internalServerId, Ram: newRam,}); e != nil {
			return e
		}
	}
	if newTrafficPackage != "" || newBillingCycle != "" {
		billingCycle := oldBillingCycle
		if newBillingCycle != "" {
			billingCycle = newBillingCycle
		}
		trafficPackage := oldTrafficPackage
		if newTrafficPackage != "" {
			trafficPackage = newTrafficPackage
		}
		if e := postServerConfigure(provider, ConfigureServerPostValues{Id: internalServerId, MonthlyPackage: trafficPackage, BillingCycle: billingCycle,}); e != nil {
			return e
		}
	}
	if newDailyBackup != "" {
		if e := postServerConfigure(provider, ConfigureServerPostValues{Id: internalServerId, DailyBackup: newDailyBackup,}); e != nil {
			return e
		}
	}
	if newManaged != "" {
		if e := postServerConfigure(provider, ConfigureServerPostValues{Id: internalServerId, Managed: newManaged,}); e != nil {
			return e
		}
	}
	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	// these values cannot be read from existing server, so we set to old values and set to new only after they were updated successfully
	newServerName := ""
	hasNewServerName := false
	if d.HasChange("name") {
		hasNewServerName = true
		o, n := d.GetChange("name")
		d.Set("name", o)
		newServerName = n.(string)
	}
	newPassword := ""
	hasNewPassword := false
	if d.HasChange("password") {
		hasNewPassword = true
		o, n := d.GetChange("password")
		d.Set("password", o)
		newPassword = n.(string)
	}
	hasNewImageId := false
	if d.HasChange("image_id") {
		hasNewImageId = true
		old, _ := d.GetChange("image_id")
		d.Set("image_id", old)
	}
	hasNewNetworks := false
	if d.HasChange("network") {
		hasNewNetworks = true
		old, _ := d.GetChange("network")
		d.Set("network", old)
	}
	hasNewSshPubKey := false
	if d.HasChange("ssh_pubkey") {
		hasNewSshPubKey = true
		old, _ := d.GetChange("ssh_pubkey")
		d.Set("ssh_pubkey", old)
	}
	if hasNewImageId {
		return errors.New("changing server image is not supported yet")
	}
	if hasNewNetworks {
		return errors.New("changing server networks is not supported yet")
	}
	if hasNewSshPubKey {
		return errors.New("changing server ssh_pubkey is not supported yet")
	}
	newCpu := ""
	newRam := ""
	oldTrafficPackage := ""
	newTrafficPackage := ""
	oldBillingCycle := ""
	newBillingCycle := ""
	if d.HasChange("server_options_id") {
		oldServerOptionsId, newServerOptionsId := d.GetChange("server_options_id")
		oldServerOptions, e := parseServerOptions(oldServerOptionsId.(string))
		if e != nil {
			return e
		}
		newServerOptions, e := parseServerOptions(newServerOptionsId.(string))
		if e != nil {
			return e
		}
		if oldServerOptions["datacenter"].(string) != newServerOptions["datacenter"].(string) {
			return errors.New("changing server datacenter is not supported yet")
		}
		if oldServerOptions["cpuType"].(string) != newServerOptions["cpuType"].(string) || oldServerOptions["cpuCores"].(string) != newServerOptions["cpuCores"].(string) {
			newCpu = fmt.Sprintf("%s%s", newServerOptions["cpuCores"].(string), newServerOptions["cpuType"].(string))
		}
		if oldServerOptions["ramMB"].(int64) != newServerOptions["ramMB"].(int64) {
			newRam = fmt.Sprintf("%v", newServerOptions["ramMB"].(int64))
		}
		oldBillingCycle = oldServerOptions["billingCycle"].(string)
		if oldServerOptions["billingCycle"].(string) != newServerOptions["billingCycle"].(string) {
			newBillingCycle = newServerOptions["billingCycle"].(string)
		}
		oldTrafficPackage = oldServerOptions["monthlyTrafficPackage"].(string)
		if oldServerOptions["monthlyTrafficPackage"].(string) != newServerOptions["monthlyTrafficPackage"].(string) {
			newTrafficPackage = newServerOptions["monthlyTrafficPackage"].(string)
		}
		if strings.Join(oldServerOptions["diskSizesGB"].([]string), ",") != strings.Join(newServerOptions["diskSizesGB"].([]string), ",") {
			return errors.New("changing server disks is not supported yet")
		}
	} else {
		oldServerOptions, e := parseServerOptions(d.Get("server_options_id").(string))
		if e != nil {
			return e
		}
		oldTrafficPackage = oldServerOptions["monthlyTrafficPackage"].(string)
		oldBillingCycle = oldServerOptions["billingCycle"].(string)
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
	provider := m.(*ProviderConfiguration)
	if e := serverConfigure(
		*provider, d.Get("internal_server_id").(string), newCpu, newRam,
		oldTrafficPackage, newTrafficPackage, oldBillingCycle, newBillingCycle,
		newDailyBackup, newManaged,
		); e != nil {
		return e
	}
	if hasNewPassword {
		if e := serverChangePassword(*provider, d.Get("internal_server_id").(string), newPassword); e != nil {
			return e
		}
		d.Set("password", newPassword)
	}
	if hasNewServerName {
		if e := serverRename(*provider, d.Get("internal_server_id").(string), newServerName); e != nil {
			return e
		}
		d.Set("name", newServerName)
	}
	if d.HasChange("power_on") {
		if d.Get("power_on").(bool) {
			if e := serverPowerOperation(*provider, d.Get("internal_server_id").(string), "poweron"); e != nil {
				return e
			}
		} else {
			if e := serverPowerOperation(*provider, d.Get("internal_server_id").(string), "poweroff"); e != nil {
				return e
			}
		}
	}
	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(*ProviderConfiguration)
	if e := serverPowerOperation(*provider, d.Get("internal_server_id").(string), "terminate"); e != nil {
		return e
	}
	return nil
}
