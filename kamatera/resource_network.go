package kamatera

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ips": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Computed: true,
			},
			"mac": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	provider := m.(*ProviderConfig)

	networks, err := getAllNetworks(provider, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	network := d.Get("network").(string)
	for _, n := range networks {
		if n.Network == network {
			d.SetId(n.Network)
			d.Set("ips", n.IPs)
			d.Set("mac", n.MAC)
			d.Set("connected", n.Connected)
		}
	}

	return diag.Errorf("no matching network")
}
