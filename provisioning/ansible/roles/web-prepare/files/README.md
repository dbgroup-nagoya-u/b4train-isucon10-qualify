# B4研修用サンプルプロジェクト（isuumo from ISUCON 10）

## 目次 <!-- omit in toc -->

- [ディレクトリ構成](#ディレクトリ構成)
- [基本的な流れ](#基本的な流れ)

## ディレクトリ構成

各ディレクトリ内に同様に`README.md`が配置してあるため，それぞれの詳細はそちらを参照してください．

- `bin`: 研修を進める上で使用する bashスクリプト群．
- `bottleneck_analysis`: ボトルネック解析の結果．
- `conf`: アプリやDBなどの設定ファイル群．
- `server_info`: マシンの基礎性能を測定する方法を記載．
- `webapp`: issumoアプリのソースコード．

## 基本的な流れ

1. 研修環境のセットアップ．
    - 各サーバ（`worker1, worker2, worker3, bench`）間でのパスワードレスSSHの設定．
    - Gitリポジトリの準備およびGitHubへの反映．
2. サーバ基礎性能の確認．
    - `server_info`内の`README.md`に従ってCPU・RAM・ストレージ・ネットワーク性能を確認．
3. アプリケーションの性能改善．
    1. ベンチマークの実行．
        - `bin`内の`run_bench.sh`を実行．
        - 例えば初期状態なら以下のコマンドで研修開始時の性能が測れる．
            ```bash
            ./bin/run_bench.sh worker1
            ```
    2. ボトルネックの確認．
        - ベンチマーク実行後に`bin`内の`post_bench.sh`を実行．
        - ボトルネックの解析結果がGitHub上に自動でpushされるため，`bottleneck_analysis`ディレクトリを確認してどのAPIが遅いか特定．
        - 例えば初期状態なら以下ようにとりあえず`main`にあげてみる．
            ```bash
            ./bin/post_bench.sh main
            ```
    3. コードの修正．
        - ボトルネックとなるAPIをアプリ（i.e., Go）実装およびDB（i.e., PostgreSQL）設計の両面から改善．
        - `conf`内の設定を変更する場合もあり．
    4. 修正の反映．
        - 修正したコードをGitHubへpush．
        - その後`bin`内の`deploy.sh`を実行することで対象サーバ全てへ変更結果を適用可能．
        - 例えば`develop`を開発用ブランチとして作成したなら以下のようになる．
            ```bash
            git push origin develop # リモートのブランチがすでに最新の状態なら不要
            ./bin/deploy.sh develop
            ```
    5. 1. に戻って計測から繰り返す．
