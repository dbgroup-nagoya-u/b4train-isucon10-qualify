# ボトルネックの分析結果

アプリケーションのボトルネック分析を行った結果（i.e., `bin/post_bench.sh`の実行結果）はこのディレクトリに配置されます．ボトルネック分析の結果は，

1. アプリケーション全体の分析結果：`nginx_summary.txt`および
2. データベースの分析結果：`db_summary.txt`

の2つに分けて出力されます．これらのファイルを見比べることでアプリケーションのどの処理がボトルネックとなっているかを特定しましょう．

例えば，`nginx_summary.txt`を見たところ`/api/estate/low_priced`の処理に時間がかかっているとわかったとします．この場合，そのAPIで使用されているSQL（以下参照）を`db_summary.txt`から探して実行時間などを見比べます．もしAPIの実行時間の大半をSQL実行で占めている場合，このSQLをなんとかして速くすることでボトルネックの解消が行えそうだということがわかります．一方でSQLの実行時間が十分小さい場合，アプリ実装（e.g., Go言語側の処理やSQLの呼び出し方）やその他の部分（e.g., ストレージIOやネットワークIO）がボトルネックだとわかります．

```sql
SELECT
  id,
  name,
  description,
  thumbnail,
  address,
  latitude,
  longitude,
  rent,
  door_height,
  door_width,
  features,
  popularity
FROM
  isuumo.estate
ORDER BY
  rent ASC,
  id ASC
LIMIT
  $1
;
```

なお，Go言語側に記述するSQLはなるべく整形しましょう．出力された`db_summary.txt`を見てもらえばわかりますが，SQLが横に長いと結果の確認に横スクロールが必要となり面倒です．そのため，基本的に以下のフォーマットで記述するのをおすすめします（例のために色々な句を全部乗せしています）．

```sql
SELECT
  column_1,
  column_2,
  ...
FROM
  table_1
  JOIN table_2
    ON join_condition_1
    AND ...
  JOIN ...
WHERE
  condition_1
  AND condition_2
  ...
GROUP BY
  column_1,
  ...
ORDER BY
  column_1 [ASC|DESC],
  ...
LIMIT
  limit_count
;
```
