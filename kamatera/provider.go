package kamatera

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ProviderConfiguration struct {
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
	}
}
