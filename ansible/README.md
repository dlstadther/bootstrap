# System Bootstrap with Ansible

## Install Ansible

Via Python:
```shell
# prepare python env, named `ansible`
pyenv install 3.8.5
pyenv global 3.8.5
python -m venv ~/venv/ansible
~/venv/ansible/bin/python -m pip install --upgrade pip

# install ansible, latest
~/venv/ansible/bin/python -m pip install ansible

# install jmespath; required for jetbrains toolbox install on Linux
~/venv/ansible/bin/pip install jmespath==0.10.0
```

## Gather Facts
```shell
~/venv/ansible/bin/ansible all -m setup
```

## Execute

### Test
```shell
~/venv/ansible/bin/ansible-playbook test.yml --inventory inventory --limit python3 --ask-become-pass
```
or
```shell
~/venv/ansible/bin/ansible-playbook test.yml --ask-become-pass
```

### install

#### LEMP9
```shell
~/venv/ansible/bin/ansible-playbook lemp9.yml --ask-become-pass
```

Only run specific tags
```shell
~/venv/ansible/bin/ansible-playbook lemp9.yml --ask-become-pass --tags "apt,zsh"
```
