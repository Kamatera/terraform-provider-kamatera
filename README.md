# Terraform Provider for Kamatera

## Installation

* [Install terraform](https://www.terraform.io/docs/index.html) (version 1.0 or higher)

## Resource Reference

* [kamatera_server resource](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/resources/server)
* [kamatera_datacenter data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/datacenter)
* [kamatera_image data source](https://registry.terraform.io/providers/Kamatera/kamatera/latest/docs/data-sources/image)

## Usage Guide

### Example Usage

This is an example end to end usage of common functionality, see sections below 
for additional functionality examples.

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
  startup_script = "echo hello from startup script > /var/hello.txt"
}
```

Init and apply:

```
terraform init && terraform apply
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

When creating a server from a private image you can't use the `kamatera_image` data resource, 
instead specify the private image name directly in the server resource `image_id` argument.

You can get the image name from Kamatare Console -> Hard Disk Library -> My Private Images

```
resource "kamatera_server" "my_server" {
  ...
  image_id = "my-private-image-name"
  ...
}
```
