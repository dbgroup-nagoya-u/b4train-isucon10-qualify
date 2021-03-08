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
