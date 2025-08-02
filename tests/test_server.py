from textwrap import dedent

import pytest

from . import common



def test_server_ram():
    with common.terraform_workdir(f'test_server_ram') as workdir:
        common.terraform_check_call(workdir, 'init')
        server_name = common.get_random_name()
        common.terraform_file(workdir, 'server.tf', dedent('''
            data "kamatera_image" "ubuntu_2404" {
              datacenter_id = "EU"
              os = "Ubuntu"
              code = "24.04 64bit"
            }
                        
            resource "kamatera_server" "my_server" {
              name = "__name__"
              datacenter_id = "EU"
              cpu_type = "B"
              cpu_cores = 2
              ram_mb = 8096
              image_id = data.kamatera_image.ubuntu_2404.id
            }
        '''), name=server_name)
        status, output = common.terraform_getstatusoutput(workdir, 'apply', '-auto-approve')
        assert status != 0, f'Expected failure due to invalid RAM size, but got success: {output}'
        assert 'invalid server configuration: unsupported RAM size for CPU type B: 8096 M' in output, f'Unexpected error message: {output}'
        common.terraform_check_call(workdir, 'destroy', '-auto-approve')
