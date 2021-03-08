# 研修環境構築手順

## 検証環境

- OS: Ubuntu 20.04 LTS
- Python: ver. 3.8.5 (OS default)
- Ansible: ver. 2.9.6

## 前提

- `root`ユーザで構築サーバへパスワードレスSSH可能．
- Ansibleがインストール済み．
    - `sudo apt install ansible`

## Playbookの実行

対象となるサーバのIPアドレスは`inventory/hosts`に指定．

### ベンチマーカ

```bash
sudo ansible-playbook bench.yaml -i inventory/hosts
```

### ワーカ

```bash
sudo ansible-playbook competitor.yaml -i inventory/hosts
```
