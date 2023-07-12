# System Bootstrap with Ansible

## Install Ansible

Via Python:
```shell
# prepare python env, named `ansible`
pyenv install 3.11.4
pyenv global 3.11.4
curl -sSL https://install.python-poetry.org | python3 -
poetry install
```

## Gather Facts
```shell
poetry run ansible all -m setup
```

## Execute

### Test
```shell
poetry run ansible-playbook test.yml --inventory inventory --limit python3 --ask-become-pass
```
or
```shell
poetry run ansible-playbook test.yml --ask-become-pass
```

### install

#### LEMP9
```shell
poetry run ansible-playbook lemp9.yml --ask-become-pass
```

Only run specific tags
```shell
poetry run ansible-playbook lemp9.yml --ask-become-pass --tags "apt,zsh"
```

#### MBP2022
```shell
poetry run ansible-playbook mbp2022.yml --ask-become-pass --check
```
