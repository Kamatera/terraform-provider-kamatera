from textwrap import dedent

import pytest

from . import common


@pytest.mark.parametrize("network_name_prefix,expected_error", [
    ('test_', 'invalid value for name (must contain only lowercase letters, digits, dashes (-) and dots (.))'),
    ('a'*21, 'expected length of name to be in the range (1 - 20)'),
    ('test-', None),
])
def test_network_name(network_name_prefix, expected_error):
    with common.terraform_workdir(f'test_network_name') as workdir:
        common.terraform_check_call(workdir, 'init')
        network_name = common.get_random_name(network_name_prefix)
        common.terraform_file(workdir, 'network.tf', dedent('''
            resource "kamatera_network" "test_network" {
              datacenter_id = "EU"
              name = "__name__"
              # a subnet with just the minimal required configuration
              subnet {
                ip = "192.168.0.0"
                bit = 23
              }
            }
        '''), name=network_name)
        status, output = common.terraform_getstatusoutput(workdir, 'apply', '-auto-approve')
        if expected_error:
            assert status != 0, 'Expected failure due to underscore in name'
            assert expected_error in output, f'Unexpected error message: {output}'
        else:
            assert status == 0, f'Expected success for valid name, but got error: {output}'
            networks_res = common.cloudcli_server_request('service/networks', json={'datacenter':'EU'})
            assert len([
                name for name
                in networks_res[0]['names']
                if name.endswith(network_name)
            ]) == 1, f'Expected exactly one network with name ending in {network_name}, but found: {networks_res}'
            common.terraform_check_call(workdir, 'destroy', '-auto-approve')
            networks_res = common.cloudcli_server_request('service/networks', json={'datacenter':'EU'})
            assert len(networks_res) == 0 or (len(networks_res) == 1 and len([
                name for name
                in networks_res[0]['names']
                if name.endswith(network_name)
            ]) == 0), f'Expected no networks with name ending in {network_name}, but found: {networks_res}'
