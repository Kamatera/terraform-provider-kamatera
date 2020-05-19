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
