# Terraform Provider for Kamatera

## Installation

* [Install terraform](https://www.terraform.io/docs/index.html) (version 1.0 or higher)

## Resource Reference

* [kamatera_server resource](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/resources/server)
* [kamatera_datacenter data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/datacenter)
* [kamatera_image data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/image)

## Usage Guide

To create a server resource, it's recommended to use our 
[server configuration interface](https://kamatera.github.io/kamateratoolbox/serverconfiggen.html?configformat=terraform) 
which provides ready to use Terraform templates with valid configuration options and identifiers according to your selection.

### Example Usage

This is an example end to end usage of common functionality, see sections below 
for additional functionality examples.

Set environment variables

```
export KAMATERA_API_CLIENT_ID=
export KAMATERA_API_SECRET=
```

Create a `main.tf` file:

```
terraform {
  required_providers {
    kamatera = {
      source = "Kamatera/kamatera"
    }
  }
}

provider "kamatera" {
}

# define the data center we will create the server and all related resources in
# see the section below "Listing available data centers" for more details
data "kamatera_datacenter" "toronto" {
  country = "Canada"
  name = "Toronto"
}

# define the server image we will create the server with
# see the section below "Listing available public images" for more details
# also see "Using a private image" if you want to use a private image you created yourself
data "kamatera_image" "ubuntu_1804" {
  datacenter_id = data.kamatera_datacenter.toronto.id
  os = "Ubuntu"
  code = "18.04 64bit"
}

# create a private network to use with the server
resource "kamatera_network" "my_private_network" {
  # the network must be in the same datacenter as the server
  datacenter_id = data.kamatera_datacenter.toronto.id
  name = "my-private-network"
  
  # define multiple subnets to use in this network
  # this subnet shows full available subnet configurations
  # the subnet below shows a more minimal example
  subnet {
    ip = "172.16.0.0"
    bit = 23
    description = "my first subnet"
    dns1 = "1.2.3.4"
    dns2 = "5.6.7.8"
    gateway = "172.16.0.100"
  }
  
  # a subnet with just the minimal required configuration
  subnet {
    ip = "192.168.0.0"
    bit = 23
  }
}

# create another private network, to show how to connect 2 networks to the server
resource "kamatera_network" "my_other_private_network" {
  datacenter_id = data.kamatera_datacenter.toronto.id
  name = "other-network"
  
  subnet {
    ip = "10.0.0.0"
    bit = 23
  }
}

# this defines the server resource with most configuration options
resource "kamatera_server" "my_server" {
  name = "my_server"
  datacenter_id = data.kamatera_datacenter.toronto.id
  cpu_type = "B"
  cpu_cores = 2
  ram_mb = 2048
  disk_sizes_gb = [15, 20]
  billing_cycle = "monthly"
  image_id = data.kamatera_image.ubuntu_1804.id
  password = "Aa123456789!"
  startup_script = "echo hello from startup script > /var/hello.txt"
  
  # this attaches a public network to the server
  # which will also allocate a public IP
  network {
    name = "wan"
  }
  
  # attach a private network with a specified ip
  network {
    # note that the network full_name attribute needs to be used
    # this value is populated with the full name of the network which may be different then 
    # the given network name
    name = resource.kamatera_network.my_private_network.full_name
    ip = "192.168.0.10"
  }
  
  # attache a private network with auto-allocated ip from the available ips in that network
  network {
    name = resource.kamatera_network.my_other_private_network.full_name
  }
}
```

Init and apply:

```
terraform init && terraform apply
```

### Listing available data centers

Add a datacenter resource without specifying any fields:

```
data "kamatera_datacenter" "frankfurt" {
}
```

Run `terraform plan`

It will output an error message containing the list of availalbe datacenters.

For example to use the Frankfurt datacenter from the following output line:

```
 │ "EU-FR"  "Germany"       "Frankfurt"  
```

The corresponding datacenter resource should look like this:

```
data "kamatera_datacenter" "frankfurt" {
  country = "Germany"
  name = "Frankfurt"
}
```

### Listing available public images

Add an image resource to your .tf file while specifying only the datacenter, for example:

```
data "kamatera_image" "my_image" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
}
```

Run `terraform plan`

It will output an error message containing the list of availalbe public images for this datacenter.

The first column in the output contains the value for the `os` argument and the second column is the `code` argument
for the image resource.

For example to use the image from the following output line:

```
│ "Ubuntu"   "22.04 64bit_optimized"        "Ubuntu Server version 22.04 LTS (Jammy Jellyfish) 64-bit.
```

The corresponding image resource should look like this:

```
data "kamatera_image" "ubuntu_1804" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  os = "Ubuntu"
  code = "22.04 64bit_optimized"
}
```

### Using a private image

You can get the private image name from Kamatare Console -> Hard Disk Library -> My Private Images

You can then use this name to specify the image data source:

```
data "kamatera_image" "my_private_image" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  private_image_name = "your-private-image-name"
}
```

This image data source can then be used the same as a public image data source in the server resource:

```
# this defines the server resource with most configuration options
resource "kamatera_server" "my_server" {
  ...
  image_id = data.kamatera_image.my_private_image.id
  ...
}
```

### Importing Existing Resources

This module supports the terraform import subcommand to import existing resources to Terraform.

#### Importing Network Resources

To get the existing network resource ID, go to Kamatera Console -> My Cloud -> Networks.
Choose the relevant datacenter. Note the datacenter ID - 2 uppercase letters.
Note the network ID - under the ID column.

The existing resource ID is `datacenter_id:network_id`

For example, to import an existing network in datacenter IL with ID 432, you would run the following:

```
terraform import kamatera_network.my_network IL:432
```

#### Importing Server Resources

To get the existing server resource ID, go to Kamatera Console -> My Cloud -> Servers.
Click on the relevant server and note the Server ID.

Example import command:

```
terraform import kamatera_server.my_server 12345678-aaaa-bbbb-cccc-1234567890ab
```
