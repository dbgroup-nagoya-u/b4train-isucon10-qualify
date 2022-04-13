# サーバの基礎性能の測定方法

サーバの基礎性能を確認ないし測定する方法を記載します．全ての確認が終わったらGitに追加して，GitHub上で全員がスペックを確認できるようにしておくと便利です．以下，全てのコマンドはこのディレクトリ上で実行する想定で書きます．

```bash
cd ~/isuumo/server_info
```

また，これらの基礎性能とは別に， **現在の** リソース使用状況を確認したい場合は`htop`が便利です．ベンチマークを実行する際は別の端末で`htop`を開いておくと，どのサーバが遊んでいるのか，逆にどのサーバに処理が集中しているのかがわかるかと思います．

```bash
htop
```

## 目次 <!-- omit in toc -->

- [CPU](#cpu)
- [RAM](#ram)
- [Storage](#storage)
- [Network](#network)

## CPU

CPUの基本的な情報は`/proc/cpuinfo`に記載されています．

```bash
cp /proc/cpuinfo ./
```

また，`lshw`を使用することでも確認できます．概ね同じ情報で，こちらの方がやや単純化されています．

```bash
sudo lshw -c cpu > ./cpu-lshw
```

## RAM

メモリの総合的な情報は`/proc/meminfo`に記載されています．

```bash
cp /proc/meminfo ./
```

どのスロットに何が挿されているかなどを知りたい場合は`lshw`で確認できます．ただし，全ての情報を出力するとやや冗長なので，以下では`-short`オプションで詳細情報を省いています．

```bash
sudo lshw -c memory -short > ./mem-lshw
```

## Storage

ストレージIOの確認には[fio](https://github.com/axboe/fio)を利用します．`fio`の実行にはワークロードなどの指定が必要ですが，サンプルの設定ファイルを用意してあるため以下のコマンドでsequential/random-read/writeでの性能が測定できます．

```bash
fio ../conf/fio/write.fio > ./storage-io-seqwrite
fio ../conf/fio/read.fio > ./storage-io-seqread
fio ../conf/fio/randwrite.fio > ./storage-io-randwrite
fio ../conf/fio/randread.fio > ./storage-io-randread
```

色々な測定結果が出てきて混乱するかと思いますが，まずは`IOPS`（IO/s）と`BW`（bandwidth）が出力されている行を見るとよいです．以下のコマンドを実行してみると，各ワークロードにおけるストレージ性能の差がよく分かるかと思います．（ISUCON 10の環境を再現するならIO bandwidthを100Mbpsくらいに制限しないといけないのですが，そっちの設定は面倒なので無視しています．どうせPostgreSQLとかOSのキャッシュとかにのるのであんまり変わらないです．）

```bash
grep 'IOPS' storage-io-*
```

## Network

ネットワークIOの測定には[iperf](https://iperf.fr/)を使用します．以降，現在ログインしているサーバが`worker1`だとして話を進めます．まず，`worker1`で以下のコマンドを実行してください．

```bash
iperf --server 1> network-io &
```

これでサーバモードの`iperf`がバックグラウンドで立ち上がり，標準出力の内容を`network-io`ファイルに出力してくれるようになりました．次に，この状態で別のワーカ（e.g., `worker2`）をクライアントとしてワーカ間のネットワークIOを計測してみます．

```bash
ssh worker2 iperf --client worker1
```

10秒ほど待つと，計測結果が標準出力に表示されます（ほぼ同じ内容が`network-io`にも書き込まれます）．次はベンチマークとワーカとの間のネットワークIOを計測してみましょう．

```bash
ssh bench iperf --client worker1
```

すべての計測が終わったら，サーバモードで立ち上げた`iperf`を終了させてください．`jobs`コマンドで起動中のプロセスを列挙できるため，`kill`コマンドで対象のプロセス（カギ括弧の中の数字の前に`%`をつけたもの）を終了させます．

```bash
jobs
kill %1 # jobsコマンドで表示された番号を入れる．他に何もプロセスを立ち上げてなければ%1．
```
