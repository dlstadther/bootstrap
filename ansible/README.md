# System Bootstrap with Ansible

## Install Ansible

Via Python:
```shell
# install uv via mise
mise install

# install dependencies
uv sync
```

## Gather Facts
```shell
uv run ansible all -m setup
```

## Execute

### Test
```shell
uv run ansible-playbook test.yml --inventory inventory --limit python3 --ask-become-pass
```
or
```shell
uv run ansible-playbook test.yml --ask-become-pass
```

### install

#### LEMP9
```shell
uv run ansible-playbook lemp9.yml --ask-become-pass
```

Only run specific tags
```shell
uv run ansible-playbook lemp9.yml --ask-become-pass --tags "apt,zsh"
```

#### MBP2022
```shell
uv run ansible-playbook mbp2022.yml --ask-become-pass --check
```

Zsh only
```shell
uv run ansible-playbook mbp2022.yml --ask-become-pass --tags "zsh"
```
