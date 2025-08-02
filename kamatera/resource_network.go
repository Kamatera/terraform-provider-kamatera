package kamatera

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
	"strings"
)

type createNetworkPostValues struct {
	Datacenter        string `json:"datacenter"`
	Name              string `json:"name"`
	SubnetIp          string `json:"subnetIp"`
	SubnetBit         int    `json:"subnetBit"`
	Gateway           string `json:"gateway"`
	Dns1              string `json:"dns1"`
	Dns2              string `json:"dns2"`
	SubnetDescription string `json:"subnetDescription"`
}

type deleteNetworkPostValues struct {
	Datacenter string `json:"datacenter"`
	Id         int    `json:"id"`
}

type createSubnetPostValues struct {
	Datacenter        string `json:"datacenter"`
	VlanId            string `json:"vlanId"`
	SubnetIp          string `json:"subnetIp"`
	SubnetBit         int    `json:"subnetBit"`
	Gateway           string `json:"gateway"`
	Dns1              string `json:"dns1"`
	Dns2              string `json:"dns2"`
	SubnetDescription string `json:"subnetDescription"`
}

type editSubnetPostValues struct {
	Datacenter        string `json:"datacenter"`
	VlanId            string `json:"vlanId"`
	SubnetId          int    `json:"subnetId"`
	SubnetIp          string `json:"subnetIp"`
	SubnetBit         int    `json:"subnetBit"`
	Gateway           string `json:"gateway"`
	Dns1              string `json:"dns1"`
	Dns2              string `json:"dns2"`
	SubnetDescription string `json:"subnetDescription"`
}

type delSubnetPostValues struct {
	SubnetId int `json:"subnetId"`
}

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceNetworkRead,
		UpdateContext: resourceNetworkUpdate,
		DeleteContext: resourceNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The network name.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 20),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9-.]+$`), "must contain only lowercase letters, digits, dashes (-) and dots (.)"),
				),
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The full network name - used internally to uniquely identify the network." +
					" This value should be used when attaching a network to a server.",
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "id attribute of datacenter data source",
			},
			"subnet": {
				Type:        schema.TypeList,
				MinItems:    0,
				MaxItems:    500,
				Optional:    true,
				Description: "IP Subnets to create and attach to this network.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The subnet IP is used with the subnet bit to determine the IP range for this subnet.",
						},
						"bit": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The subnet bit is used with the subnt IP to determine the IP range for this subnet.",
						},
						"gateway": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional gateway IP from within the subnet IP range.",
						},
						"dns1": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional primary DNS server IP for this subnet.",
						},
						"dns2": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional secondary DNS server IP for this subnet.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional description of this subnet.",
						},
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The unique subnet ID.",
						},
					},
				},
			},
			"network_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)
	subnets := d.Get("subnet").([]interface{})
	if len(subnets) < 1 {
		return diag.Errorf("when creating a new network, at least 1 subnet is required")
	}
	subnetDescs := make(map[string]bool)
	for _, subnet := range subnets {
		subnetDescs[subnet.(map[string]interface{})["description"].(string)] = true
	}
	if len(subnetDescs) != len(subnets) {
		return diag.Errorf("each subnet must have a unique description")
	}
	firstSubnet := subnets[0].(map[string]interface{})
	body := &createNetworkPostValues{
		Datacenter:        d.Get("datacenter_id").(string),
		Name:              d.Get("name").(string),
		SubnetIp:          firstSubnet["ip"].(string),
		SubnetBit:         firstSubnet["bit"].(int),
		Gateway:           firstSubnet["gateway"].(string),
		Dns1:              firstSubnet["dns1"].(string),
		Dns2:              firstSubnet["dns2"].(string),
		SubnetDescription: firstSubnet["description"].(string),
	}
	result, err := request(provider, "POST", "service/network/create", body)
	if err != nil {
		return diag.FromErr(err)
	}
	response := result.(map[string]interface{})
	var res map[string]interface{}
	err = json.Unmarshal([]byte(response["res"].(string)), &res)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", res["networkId"].(float64)))
	firstSubnet["id"] = res["subnetId"].(float64)
	for _, subnet := range subnets {
		if subnet.(map[string]interface{})["description"].(string) != firstSubnet["description"] {
			_, err := addSubnet(provider, d, subnet.(map[string]interface{}))
			if err != nil {
				return err
			}
		}
	}
	d.Set("subnet", subnets)
	return resourceNetworkRead(ctx, d, m)
}

func resourceNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)
	datacenter := d.Get("datacenter_id").(string)
	id := d.Id()
	result, err := request(provider, "GET", fmt.Sprintf("service/networks?datacenter=%s", datacenter), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	var network map[string]interface{}
	networks := result.([]interface{})
	for _, network_ := range networks {
		network__ := network_.(map[string]interface{})
		if fmt.Sprintf("%v", network__["vlanId"].(float64)) == id {
			network = network__
			break
		}
	}
	if network == nil {
		return diag.Errorf("Did not find network %v in datacenter %s", id, datacenter)
	}
	networkIds := network["ids"].([]interface{})
	if len(networkIds) != 1 {
		return diag.Errorf("Invalid ids returned from network list")
	}
	networkNames := network["names"].([]interface{})
	if len(networkNames) != 1 {
		return diag.Errorf("Invalid names returned from network list")
	}
	d.Set("network_id", networkIds[0].(float64))
	fullName := networkNames[0].(string)
	d.Set("full_name", fullName)
	if d.Get("name").(string) == "" {
		fullNameParts := strings.Split(fullName, "-")
		d.Set("name", strings.Join(fullNameParts[2:], "-"))
	}

	subnetsResult, err := request(provider, "GET", fmt.Sprintf("service/network/subnets?datacenter=%s&vlanId=%s", datacenter, id), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	var subnets []map[string]interface{}
	subnetDescs := make(map[string]bool)
	for _, subnet_ := range subnetsResult.([]interface{}) {
		subnet__ := subnet_.(map[string]interface{})
		description := subnet__["subnetDescription"].(string)
		subnets = append(subnets, map[string]interface{}{
			"ip":          subnet__["subnetIp"].(string),
			"bit":         subnet__["subnetBit"].(float64),
			"gateway":     subnet__["gateway"].(string),
			"dns1":        subnet__["dns1"].(string),
			"dns2":        subnet__["dns2"].(string),
			"description": description,
			"id":          subnet__["subnetId"].(float64),
		})
		subnetDescs[description] = true
	}
	if len(subnetDescs) != len(subnets) {
		return diag.Errorf("Invalid subnet description - cannot differentiate between subnets based on subnet descriptions")
	}
	d.Set("subnet", subnets)
	return
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)
	if d.HasChange("name") {
		return diag.Errorf("changing network name is not supported")
	}
	if d.HasChange("datacenter_id") {
		return diag.Errorf("changing network datacenter is not supported")
	}
	if d.HasChange("subnet") {
		oldSubnets, newSubnets := d.GetChange("subnet")
		newSubnetsByDescription := make(map[string]map[string]interface{})
		oldSubnetsByDescription := make(map[string]map[string]interface{})
		for _, newSubnet := range newSubnets.([]interface{}) {
			newSubnet := newSubnet.(map[string]interface{})
			newSubnetsByDescription[newSubnet["description"].(string)] = newSubnet
		}
		for _, oldSubnet := range oldSubnets.([]interface{}) {
			oldSubnet := oldSubnet.(map[string]interface{})
			oldSubnetsByDescription[oldSubnet["description"].(string)] = oldSubnet
		}
		if len(oldSubnets.([]interface{})) != len(oldSubnetsByDescription) || len(newSubnets.([]interface{})) != len(newSubnetsByDescription) {
			return diag.Errorf("Invalid subnet descriptions - cannot identify unique subnets based on descriptions")
		}
		for description, newSubnet := range newSubnetsByDescription {
			oldSubnet, oldExists := oldSubnetsByDescription[description]
			if oldExists {
				if isSubnetDifferent(oldSubnet, newSubnet) {
					err := editSubnet(provider, d, newSubnet)
					if err != nil {
						return err
					}
				}
			} else {
				newSubnetId, err := addSubnet(provider, d, newSubnet)
				if err != nil {
					return err
				}
				newSubnet["id"] = newSubnetId
			}
		}
		for description, oldSubnet := range oldSubnetsByDescription {
			_, newExists := newSubnetsByDescription[description]
			if !newExists {
				err := delSubnet(provider, d, oldSubnet)
				if err != nil {
					return err
				}
			}
		}
	}
	return resourceNetworkRead(ctx, d, m)
}

func resourceNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*ProviderConfig)
	for _, subnet := range d.Get("subnet").([]interface{}) {
		err := delSubnet(provider, d, subnet.(map[string]interface{}))
		if err != nil {
			return err
		}
	}
	body := &deleteNetworkPostValues{
		Datacenter: d.Get("datacenter_id").(string),
		Id:         d.Get("network_id").(int),
	}
	_, err := request(provider, "POST", "service/network/delete", body)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceNetworkImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	d.Set("datacenter_id", idParts[0])
	d.SetId(idParts[1])
	diags := resourceNetworkRead(ctx, d, m)
	if diags.HasError() {
		var errorMessages []string
		for i := range diags {
			errorMessages = append(errorMessages, diags[i].Summary)
		}
		return nil, fmt.Errorf(strings.Join(errorMessages, ", "))
	}
	return []*schema.ResourceData{d}, nil
}

func isSubnetDifferent(subnet1 map[string]interface{}, subnet2 map[string]interface{}) bool {
	return subnet1["ip"].(string) != subnet2["ip"].(string) ||
		subnet1["bit"].(int) != subnet2["bit"].(int) ||
		subnet1["gateway"].(string) != subnet2["gateway"].(string) ||
		subnet1["dns1"].(string) != subnet2["dns1"].(string) ||
		subnet1["dns2"].(string) != subnet2["dns2"].(string)
}

func editSubnet(provider *ProviderConfig, d *schema.ResourceData, subnet map[string]interface{}) diag.Diagnostics {
	body := &editSubnetPostValues{
		Datacenter:        d.Get("datacenter_id").(string),
		VlanId:            d.Id(),
		SubnetId:          subnet["id"].(int),
		SubnetIp:          subnet["ip"].(string),
		SubnetBit:         subnet["bit"].(int),
		Gateway:           subnet["gateway"].(string),
		Dns1:              subnet["dns1"].(string),
		Dns2:              subnet["dns2"].(string),
		SubnetDescription: subnet["description"].(string),
	}
	_, err := request(provider, "POST", "service/network/subnet/edit", body)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func delSubnet(provider *ProviderConfig, d *schema.ResourceData, subnet map[string]interface{}) diag.Diagnostics {
	body := &delSubnetPostValues{
		SubnetId: subnet["id"].(int),
	}
	_, err := request(provider, "POST", "service/network/subnet/delete", body)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func addSubnet(provider *ProviderConfig, d *schema.ResourceData, subnet map[string]interface{}) (float64, diag.Diagnostics) {
	body := &createSubnetPostValues{
		Datacenter:        d.Get("datacenter_id").(string),
		VlanId:            d.Id(),
		SubnetIp:          subnet["ip"].(string),
		SubnetBit:         subnet["bit"].(int),
		Gateway:           subnet["gateway"].(string),
		Dns1:              subnet["dns1"].(string),
		Dns2:              subnet["dns2"].(string),
		SubnetDescription: subnet["description"].(string),
	}
	result, err := request(provider, "POST", "service/network/subnet/create", body)
	if err != nil {
		return 0, diag.FromErr(err)
	}
	response := result.(map[string]interface{})
	var res map[string]interface{}
	err = json.Unmarshal([]byte(response["res"].(string)), &res)
	if err != nil {
		return 0, diag.FromErr(err)
	}
	return res["subnetId"].(float64), nil
}
