import os
import datetime
import shutil
import tempfile
import subprocess
from contextlib import contextmanager

import requests

KAMATERA_API_URL = os.environ.get('KAMATERA_API_URL', 'https://cloudcli.cloudwm.com')
KAMATERA_API_CLIENT_ID = os.environ.get('KAMATERA_API_CLIENT_ID')
KAMATERA_API_SECRET = os.environ.get('KAMATERA_API_SECRET')

MAIN_TF_TEMPLATE = '''
terraform {
  required_providers {
    kamatera = {
      source = "Kamatera/kamatera"
    }
  }
}

provider "kamatera" {
}
'''


def get_random_name(prefix='', suffix=''):
    import random
    import string
    name = ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))
    return f"{prefix}{name}{suffix}"


@contextmanager
def terraform_workdir(prefix):
    basedir = os.path.join(os.path.dirname(__file__), '..', '.data', 'tests_workdirs')
    os.makedirs(basedir, exist_ok=True)
    workdir = tempfile.mkdtemp(prefix=prefix + '_' + datetime.datetime.now().strftime('%Y%m%d%H%S') + '_', dir=basedir)
    with open(os.path.join(workdir, 'main.tf'), 'w') as f:
        f.write(MAIN_TF_TEMPLATE)
    yield workdir
    shutil.rmtree(workdir, ignore_errors=True)


def terraform_file(workdir, filename, content, **kwargs):
    for k, v in kwargs.items():
        content = content.replace(f'__{k}__', v)
    filepath = os.path.join(workdir, filename)
    with open(filepath, 'w') as f:
        f.write(content)


def terraform_check_call(workdir, *cmd):
    subprocess.check_call(['terraform', f'-chdir={workdir}', *cmd])


def terraform_getstatusoutput(workdir, *cmd):
    return subprocess.getstatusoutput(' '.join(['terraform', f'-chdir={workdir}', *cmd]))


def cloudcli_server_request(path, **kwargs):
    url = os.path.join(KAMATERA_API_URL, path)
    method = kwargs.pop("method", "GET")
    res = requests.request(method=method, url=url, headers={
        "AuthClientId": os.environ["KAMATERA_API_CLIENT_ID"],
        "AuthSecret": os.environ["KAMATERA_API_SECRET"],
        "Content-Type": "application/json",
        "Accept": "application/json"
    }, **kwargs)
    if res.status_code != 200:
        raise Exception(res.json())
    else:
        return res.json()
