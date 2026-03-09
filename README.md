# macos_setup

1. Install brew
2. `$ brew bundle`
3. Check apps.md for remaining applications/services

## Ansible

Requires [mise](https://mise.jdx.dev). From the `ansible/` directory:

```shell
cd ansible
mise install   # installs uv
uv sync        # installs project dependencies
```


Copy Zsh only
```shell
cp -r ./ansible/roles/workstations/files/zsh/ ~/
```
