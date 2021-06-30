package kamatera

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ProviderConfig struct {
	ApiUrl      string
	ApiClientID string
	ApiSecret   string
}

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kamatera_server":         resourceServer(),
			"kamatera_server_network": resourceNetwork(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"kamatera_datacenter": dataSourceDatacenter(),
			"kamatera_image":      dataSourceImage(),
		},
		Schema: map[string]*schema.Schema{
			"api_client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_CLIENT_ID", nil),
				Description: "Kamatera API Client ID",
			},
			"api_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_SECRET", nil),
				Description: "Kamatera API Secret",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_URL", "https://cloudcli.cloudwm.com"),
				Description: "Kamatera API Url",
			},
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiClientID := d.Get("api_client_id").(string)
	apiSecret := d.Get("api_secret").(string)
	apiURL := d.Get("api_url").(string)

	return &ProviderConfig{
		ApiUrl:      apiURL,
		ApiClientID: apiClientID,
		ApiSecret:   apiSecret,
	}, nil
}
