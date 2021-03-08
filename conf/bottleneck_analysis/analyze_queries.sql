SELECT
'==================================================
Order by call counts
==================================================' AS sort_type;
SELECT * FROM analyze_slow_queries('calls', 0);
SELECT * FROM analyze_slow_queries('calls', 10);

SELECT
'==================================================
Order by total time
==================================================' AS sort_type;
SELECT * FROM analyze_slow_queries('total_time', 0);
SELECT * FROM analyze_slow_queries('total_time', 10);

SELECT
'==================================================
Order by mean time
==================================================' AS sort_type;
SELECT * FROM analyze_slow_queries('mean_time', 0);
SELECT * FROM analyze_slow_queries('mean_time', 10);

SELECT
'==================================================
Order by standard deviation time
==================================================' AS sort_type;
SELECT * FROM analyze_slow_queries('stddev_time', 0);
SELECT * FROM analyze_slow_queries('stddev_time', 10);

SELECT
'==================================================
Order by max time
==================================================' AS sort_type;
SELECT * FROM analyze_slow_queries('max_time', 0);
SELECT * FROM analyze_slow_queries('max_time', 10);
