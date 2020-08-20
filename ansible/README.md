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

