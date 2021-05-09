package kamatera

import (
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
		ResourcesMap:   map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"server": dataSourceServer(),
		},
		Schema: map[string]*schema.Schema{
			"api_client_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_CLIENT_ID", nil),
				Description: "Kamatera API Client ID",
			},
			"api_secret": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_SECRET", nil),
				Description: "Kamatera API Secret",
			},
			"api_url": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("KAMATERA_API_URL", "https://cloudcli.cloudwm.com"),
				Description: "Kamatera API Url",
			},
		},
	}
}
