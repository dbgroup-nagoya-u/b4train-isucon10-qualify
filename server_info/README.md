# サーバ情報

## CPU

```bash
cp /proc/cpuinfo ./
```

## RAM

```bash
cp /proc/meminfo ./
```

## Storage

```bash
fio ../conf/fio/write.fio > ./fio_write
fio ../conf/fio/read.fio > ./fio_read
fio ../conf/fio/randwrite.fio > ./fio_randwrite
fio ../conf/fio/randread.fio > ./fio_randread
```

## Network

以下`worker1`をサーバ，`worker2`をクライアントとして動作させる想定．

サーバ側．クライアント側の実行が終了したら`Ctrl+C`．

```bash
iperf -s
```

クライアント側．

```bash
iperf -c worker1 > ./iperf_client
```
