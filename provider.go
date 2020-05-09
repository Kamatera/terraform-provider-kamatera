package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ConfigureFunc: ConfigureProvider,

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
		DataSourcesMap: map[string]*schema.Resource{
			"kamatera_datacenter": dataSourceDatacenter(),
			"kamatera_image": dataSourceImage(),
			"kamatera_server_options": dataSourceServerOptions(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"kamatera_server": resourceServer(),
		},
	}
}

type ProviderConfiguration struct {
	ApiUrl      string
	ApiClientID string
	ApiSecret   string
}

func ConfigureProvider(d *schema.ResourceData) (interface{}, error) {
	config := ProviderConfiguration{
		ApiUrl:      d.Get("api_url").(string),
		ApiClientID: d.Get("api_client_id").(string),
		ApiSecret:   d.Get("api_secret").(string),
	}
	return &config, nil
}
