# Terraform Provider for Kamatera

## Installation

* [Install terraform](https://www.terraform.io/docs/index.html) (version 1.0 or higher)

## Resource Reference

* [kamatera_server resource](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/resources/server)
* [kamatera_datacenter data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/datacenter)
* [kamatera_image data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/image)

## Example Usage

Set environment variables

```
export KAMATERA_API_CLIENT_ID=
export KAMATERA_API_SECRET=
```

Create a `main.tf` file, for example (replace server `name` and `password`):

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

data "kamatera_datacenter" "petach_tikva" {
  country = "Israel"
  name = "Petach Tikva"
}

data "kamatera_image" "ubuntu_1804" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  os = "Ubuntu"
  code = "18.04 64bit"
}

resource "kamatera_server" "my_server" {
  name = "my_server"
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  cpu_type = "B"
  cpu_cores = 2
  ram_mb = 2048
  disk_sizes_gb = [15, 20]
  billing_cycle = "monthly"
  image_id = data.kamatera_image.ubuntu_1804.id
  password = "Aa123456789!"
}
```

Init and apply:

```
terraform init && terraform apply
```
