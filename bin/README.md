# ISUCON用スクリプト

## bootstrap.sh

### 使用方法

```bash
./bin/bootstrap.sh
```

### 用途

各サーバの初期設定を行う．設定内容の詳細はスクリプトを参照．

## deploy.sh

### 使用方法

```bash
./bin/deploy.sh <git_branch_name>
```

### 用途

ソースコードへ修正を加えた後，ベンチマーク実施前に修正内容をサーバへ反映する．

### 前提

- 適用対象のサーバへパスワードレスのSSH接続が可能．
- 各サーバのワークスペースに以下のファイルが存在．
    - `bin/deploy.sh`
    - `bin/component/fetch_branch.sh`

## run_bench.sh

### 使用方法

```bash
./bin/run_bench.sh <target_host>
```

### 用途

指定したサーバに対してベンチマークを実行する．

### 前提

- 対象ホストでWebサーバが動いている．
- スクリプトの都合上，ホスト名は`worker1`，`worker2`，`worker3`で設定すること．

## post_bench.sh

### 使用方法

```bash
./bin/post_bench.sh <git_branch_name>
```

### 用途

ベンチマーク実施後，アプリ・DBのボトルネック分析を行い，分析結果を指定されたブランチへプッシュする．

### 前提

- 適用対象のサーバへパスワードレスのSSH接続が可能．
