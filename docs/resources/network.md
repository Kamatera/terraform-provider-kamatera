---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kamatera_network Resource - terraform-provider-kamatera"
subcategory: ""
description: |-
  
---

# kamatera_network (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **datacenter_id** (String) id attribute of datacenter data source
- **name** (String) The network name.

### Optional

- **id** (String) The ID of this resource.
- **subnet** (Block List, Max: 500) IP Subnets to create and attach to this network. (see [below for nested schema](#nestedblock--subnet))

### Read-Only

- **full_name** (String) The full network name - used internally to uniquely identify the network. This value should be used when attaching a network to a server.
- **network_id** (Number)

<a id="nestedblock--subnet"></a>
### Nested Schema for `subnet`

Required:

- **bit** (Number) The subnet bit is used with the subnt IP to determine the IP range for this subnet.
- **ip** (String) The subnet IP is used with the subnet bit to determine the IP range for this subnet.

Optional:

- **description** (String) Optional description of this subnet.
- **dns1** (String) Optional primary DNS server IP for this subnet.
- **dns2** (String) Optional secondary DNS server IP for this subnet.
- **gateway** (String) Optional gateway IP from within the subnet IP range.

Read-Only:

- **id** (Number) The unique subnet ID.


