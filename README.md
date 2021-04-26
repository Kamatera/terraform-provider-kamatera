# Terraform Provider for Kamatera

## Installation

* [Install terraform](https://www.terraform.io/docs/index.html) (version 0.12 or higher)
* Download the Kamatera provider for your OS/architecture from [latest release](https://github.com/Kamatera/terraform-provider-kamatera/releases)
* Unzip and place the binary in your PATH named `terraform-provider-kamatera`, for example:
  * `unzip terraform-provider-kamatera-0.0.4-linux-amd64.zip`
  * `sudo mv terraform-provider-kamatera-0.0.4-linux-amd64 /usr/local/bin/terraform-provider-kamatera`

## Usage

Set environment variables

```
export KAMATERA_API_CLIENT_ID=
export KAMATERA_API_SECRET=
```

Create a `main.tf` file, for example:
- 0.12 and earlier:

```
provider "kamatera" {}

data "kamatera_datacenter" "petach_tikva" {
  country = "Israel"
  name = "Petach Tikva"
}

data "kamatera_image" "ubuntu_1804" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  os = "Ubuntu"
  code = "18.04 64bit"
}

data "kamatera_server_options" "B2_2048_monthly" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  cpu_type = "B"
  cpu_cores = 2
  ram_mb = 2048
  disk_size_gb = 15
  extra_disk_sizes_gb = [20, 30]
  billing_cycle = "monthly"
}

resource "kamatera_server" "terraformtest" {
  name = "terraformtest"
  server_options_id = data.kamatera_server_options.B2_2048_monthly.id
  image_id = data.kamatera_image.ubuntu_1804.id
}
```

- 0.13 and later:

```
terraform {
  required_providers {
    kamatera = {
      source = "kamatera/kamatera"
      version = "0.0.4"
    }
  }
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

data "kamatera_server_options" "B2_2048_monthly" {
  datacenter_id = data.kamatera_datacenter.petach_tikva.id
  cpu_type = "B"
  cpu_cores = 2
  ram_mb = 2048
  disk_size_gb = 15
  extra_disk_sizes_gb = [20, 30]
  billing_cycle = "monthly"
}

resource "kamatera_server" "terraformtest" {
  name = "terraformtest"
  server_options_id = data.kamatera_server_options.B2_2048_monthly.id
  image_id = data.kamatera_image.ubuntu_1804.id
}
```

Init and apply:

```
terraform init && terraform apply
```
