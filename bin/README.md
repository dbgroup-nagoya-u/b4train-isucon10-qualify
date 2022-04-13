# B4研修用スクリプト

## 目次 <!-- omit in toc -->

- [deploy.sh](#deploysh)
- [run_bench.sh](#run_benchsh)
- [post_bench.sh](#post_benchsh)
- [bootstrap.sh](#bootstrapsh)

## deploy.sh

指定したブランチのソースコードを各サーバへ反映します．

### 使用方法

```bash
./bin/deploy.sh <git_branch_name>
```

### 前提

- 適用対象のサーバへパスワードレスのSSH接続が可能．
- 各サーバのワークスペースに以下のファイルが存在．
    - `bin/deploy.sh`
    - `bin/component/fetch_branch.sh`

## run_bench.sh

### 使用方法

指定したサーバに対してベンチマークを実行します．

```bash
./bin/run_bench.sh <target_host>
```

### 前提

- 対象ホストでWebサーバが動いている．
- スクリプトの都合上，ホスト名は`worker1`，`worker2`，`worker3`で設定すること．

## post_bench.sh

ベンチマークの実行ログからアプリ・DBのボトルネック分析を行い，分析結果を指定されたブランチへpushします．

### 使用方法

```bash
./bin/post_bench.sh <git_branch_name>
```

### 前提

- 適用対象のサーバへパスワードレスのSSH接続が可能．

## bootstrap.sh

研修用サーバの初期設定を行うスクリプトですが，すでに適用済みのため研修時には実行しなくてOKです．普通のISUCONでも（だいたい）使えるように組んであるため，よければ参加時は使用してください．

### 使用方法

```bash
./bin/bootstrap.sh
```
