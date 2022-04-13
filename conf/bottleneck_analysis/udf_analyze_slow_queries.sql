DROP FUNCTION IF EXISTS analyze_slow_queries;

CREATE FUNCTION analyze_slow_queries (
  order_by TEXT,
  offset_size INT4
) RETURNS TABLE (
  rank INT8,
  ____________________________query____________________________ TEXT,
  calls INT8,
  total TEXT,
  mean TEXT,
  stddev TEXT,
  min TEXT,
  max TEXT,
  rows INT8,
  shared_blks_hit INT8,
  shared_blks_read INT8,
  shared_blks_dirtied INT8,
  shared_blks_written INT8,
  local_blks_hit INT8,
  local_blks_read INT8,
  local_blks_dirtied INT8,
  local_blks_written INT8,
  temp_blks_read INT8,
  temp_blks_written INT8
)
IMMUTABLE AS $$
DECLARE
BEGIN
  RETURN QUERY
    SELECT
      ROW_NUMBER() OVER (
        ORDER BY
          CASE
            WHEN order_by = 'calls' THEN stat.calls
            WHEN order_by = 'total_time' THEN stat.total_time
            WHEN order_by = 'mean_time' THEN stat.mean_time
            WHEN order_by = 'stddev_time' THEN stat.stddev_time
            ELSE stat.max_time
          END
        DESC
      ),
      regexp_replace(stat.query, '\([, $0-9()"a-z_]+\)','(*)'),
      stat.calls,
      to_char(stat.total_time, '99999.999'),
      to_char(stat.mean_time, '9999.999'),
      to_char(stat.stddev_time, '9999.999'),
      to_char(stat.min_time, '9999.999'),
      to_char(stat.max_time, '9999.999'),
      stat.rows,
      stat.shared_blks_hit,
      stat.shared_blks_read,
      stat.shared_blks_dirtied,
      stat.shared_blks_written,
      stat.local_blks_hit,
      stat.local_blks_read,
      stat.local_blks_dirtied,
      stat.local_blks_written,
      stat.temp_blks_read,
      stat.temp_blks_written
    FROM
      pg_stat_statements AS stat
    WHERE
      NOT stat.query LIKE '%pg_catalog%'
      AND NOT stat.query LIKE '%pg_stat%'
      AND NOT stat.query LIKE '%public.%'
      AND NOT stat.query LIKE '%features%'
      AND NOT stat.query LIKE 'BEGIN%'
      AND NOT stat.query LIKE 'COMMIT'
      AND NOT stat.query LIKE 'ROLLBACK'
    LIMIT
      10
    OFFSET
      offset_size;
  RETURN;
END;
$$ LANGUAGE plpgsql;
