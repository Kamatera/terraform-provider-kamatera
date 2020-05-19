#!/usr/bin/env python3
import os
import subprocess
import json


os.makedirs("tests/output", exist_ok=True)
CREATE_SERVER_NAME = "terraformtest"

CLOUDCLI_ARGS = ["--api-clientid", os.environ["KAMATERA_API_CLIENT_ID"], "--api-secret", os.environ["KAMATERA_API_SECRET"]]


def test_create_server():
    with open("tests/output/main.tf", "w") as f:
        f.write("""
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

    resource "kamatera_server" "__CREATE_SERVER_NAME__" {
      name = "__CREATE_SERVER_NAME__"
      server_options_id = data.kamatera_server_options.B2_2048_monthly.id
      image_id = data.kamatera_image.ubuntu_1804.id
    }
    """.replace("__CREATE_SERVER_NAME__", CREATE_SERVER_NAME))
    if os.path.exists("terraform.tfstate") or os.path.exists("terraform.tfstate.backup"):
        raise Exception("Existing terraform state")
    subprocess.check_call(["terraform", "init", "tests/output"])
    print("Creating server...")
    subprocess.check_call(["terraform", "apply", "-auto-approve", "-input=false", "tests/output"])
    with open("tests/output/test_create.json", "wb") as f:
        f.write(subprocess.check_output(["cloudcli", *CLOUDCLI_ARGS, "server", "info", "--format", "json", "--name", "%s.*" % CREATE_SERVER_NAME]))
    with open("tests/output/test_create.json", "rb") as f:
        servers_info = json.load(f)
    assert len(servers_info) == 1
    server = servers_info[0]
    assert len(server["id"]) > 5
    assert server["datacenter"] == "IL-PT"
    assert server["cpu"] == "2B"
    assert server["name"].startswith(CREATE_SERVER_NAME)
    assert server["ram"] == 2048
    assert server["power"] == "on"
    assert server["diskSizes"] == [15, 20, 30]
    assert server["networks"][0]["network"] == "wan-il-pt"
    assert server["billing"] == "monthly"
    assert server["traffic"] == "t5000"
    assert server["managed"] == "0"
    assert server["backup"] == "0"


def test_stop_server():
    with open("tests/output/main.tf", "w") as f:
        f.write("""
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

    resource "kamatera_server" "__CREATE_SERVER_NAME__" {
      name = "__CREATE_SERVER_NAME__"
      server_options_id = data.kamatera_server_options.B2_2048_monthly.id
      image_id = data.kamatera_image.ubuntu_1804.id
      power_on = false
    }
    """.replace("__CREATE_SERVER_NAME__", CREATE_SERVER_NAME))
    print("Stopping server...")
    subprocess.check_call(["terraform", "apply", "-auto-approve", "-input=false", "tests/output"])
    with open("tests/output/test_stop.json", "wb") as f:
        f.write(subprocess.check_output(["cloudcli", *CLOUDCLI_ARGS, "server", "info", "--format", "json", "--name", "%s.*" % CREATE_SERVER_NAME]))
    with open("tests/output/test_stop.json", "rb") as f:
        servers_info = json.load(f)
    assert len(servers_info) == 1
    server = servers_info[0]
    assert server["name"].startswith(CREATE_SERVER_NAME)
    assert server["power"] == "off"


def test_change_server_options():
    with open("tests/output/main.tf", "w") as f:
        f.write("""
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

    data "kamatera_server_options" "B1_1024_monthly" {
      datacenter_id = data.kamatera_datacenter.petach_tikva.id
      cpu_type = "B"
      cpu_cores = 1
      ram_mb = 1024
      disk_size_gb = 15
      extra_disk_sizes_gb = [20, 30]
      billing_cycle = "monthly"
    }

    resource "kamatera_server" "__CREATE_SERVER_NAME__" {
      name = "__CREATE_SERVER_NAME__"
      server_options_id = data.kamatera_server_options.B1_1024_monthly.id
      image_id = data.kamatera_image.ubuntu_1804.id
      power_on = false
    }
    """.replace("__CREATE_SERVER_NAME__", CREATE_SERVER_NAME))
    print("Changing server options...")
    subprocess.check_call(["terraform", "apply", "-auto-approve", "-input=false", "tests/output"])
    with open("tests/output/test_change_options.json", "wb") as f:
        f.write(subprocess.check_output(["cloudcli", *CLOUDCLI_ARGS, "server", "info", "--format", "json", "--name", "%s.*" % CREATE_SERVER_NAME]))
    with open("tests/output/test_change_options.json", "rb") as f:
        servers_info = json.load(f)
    assert len(servers_info) == 1
    server = servers_info[0]
    assert server["cpu"] == "1B"
    assert server["name"].startswith(CREATE_SERVER_NAME)
    assert server["ram"] == 1024


def test_destroy_server():
    print("Destroying server...")
    subprocess.check_call(["terraform", "destroy", "-auto-approve", "tests/output"])
    status, output = subprocess.getstatusoutput("cloudcli %s server info --format json --name \"%s.*\"" % (" ".join(CLOUDCLI_ARGS), CREATE_SERVER_NAME))
    assert status == 2, output
    assert "No servers found" in output, output


test_create_server()
test_stop_server()
test_change_server_options()
test_destroy_server()
print("Great Success!")
