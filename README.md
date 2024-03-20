# 名古屋大学DBグループB4研修

[ISUCON](https://isucon.net/)の第10回予選問題に変更を加えた，研究室に配属された新B4のための研修プロジェクトです．

## ディレクトリ構成

- `bin`: 問題解答時に利用すると便利なスクリプト群．
- `bottleneck_analysis`: スクリプトによって自動生成されるボトルネックの解析結果．
- `conf`: 各種設定ファイル．スクリプトによって各サーバへ反映．
- `server_info`: 実施環境の基本的な性能測定結果．
- `webapp`: アプリの実装コード．

## 想定環境

| 項目 | 値 |
|:---|:---|
| OS | Ubuntu 22.04 LTS |
| DB | PostgreSQL 14 |
| Webサーバ | Nginx |
| アプリ実装言語 | Go 1.21.8 |

### サービス稼働用サーバスペック

| 項目 | 値 |
|:---|:---|
| CPU | 1 core |
| RAM | 2 GB |
| Network IO (LAN) | 1 Gbps |
| Network IO (WAN) | 100 Mbps |


### ベンチマーク用サーバスペック

| 項目 | 値 |
|:---|:---|
| CPU | 2 core |
| RAM | 8 GB |
| Network IO (WAN) | 100 Mbps |

## 参考リンク

- [ISUCON10 予選リポジトリ](https://github.com/isucon/isucon10-qualify)
- [ISUCON10 予選レギュレーション](http://isucon.net/archives/54753430.html)
- [ISUCON10 予選当日マニュアル](https://gist.github.com/progfay/25edb2a9ede4ca478cb3e2422f1f12f6)

## 謝辞

本リポジトリはISUCONによって公開されている上記リポジトリを基に作成しています．ここに記し，感謝の意を表します．

## 注釈

「[ISUCON](https://isucon.net/)」は，LINE株式会社の商標または登録商標です．
