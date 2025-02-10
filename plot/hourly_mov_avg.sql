--- calculate hourly moving average of `shielded vs unshielded` inclusion rate. save to output file `hourly.csv`

\o hourly.csv
\pset fieldsep ','

WITH hourly_counts AS (
    SELECT
        date_trunc('hour', created_at) AS hour,
        SUM(CASE WHEN tx_status = 'shielded inclusion' THEN 1 ELSE 0 END) AS shielded_count,
        SUM(CASE WHEN tx_status = 'unshielded inclusion' THEN 1 ELSE 0 END) AS unshielded_count,
	COUNT(*) as total_count
    FROM
        decrypted_tx
    WHERE
	created_at > 'now'::timestamp - '7 day'::interval
    GROUP BY
        hour
),
moving_average AS (
    SELECT
        hour,
        COALESCE(shielded_count::float / NULLIF(total_count, 0), 0) AS shielded_percentage,
	total_count
    FROM
        hourly_counts
),
filled_intervals AS (
    SELECT
        generate_series(
            (SELECT MIN(hour) FROM hourly_counts),
            (SELECT MAX(hour) FROM hourly_counts),
            '1 hour'::interval
        ) AS hour
)
SELECT
    fi.hour,
    COALESCE(ma.shielded_percentage, LAG(ma.shielded_percentage) OVER (ORDER BY fi.hour)) AS moving_average,
    total_count
FROM
    filled_intervals fi
LEFT JOIN
    moving_average ma ON fi.hour = ma.hour
ORDER BY
    fi.hour;

