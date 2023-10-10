```
CREATE TABLE api_logs
(
    api_timestamp DateTime,
    completion_time_ms UInt32,
    url String,
    http_status_code UInt16,
    third_party_name String,
    tenantId UUID,
    entity_type String,
    entity_id String,
    operation_category String,
    operation_subcategory String
) ENGINE = MergeTree()
PARTITION BY tenantId
// probably remove tenantID here, since we’re partitioning by tenant ID.
ORDER BY (tenantId, api_timestamp);
```
# Engine choice

mergeTree: real-time analytics with high-speed insertions.

AggregatingMergeTree: store aggregate function states and is useful when raw data is vast, but its aggregated view is comparatively small.

ReplacingMergeTree: Useful when you periodically get full snapshots of data and want to replace older data with newer ones based on a primary key.

MaterializedView: Allows storing the result of a query and can be tied to another table to update automatically when new data is added.

# Updating/Deleting data

https://clickhouse.com/blog/handling-updates-and-deletes-in-clickhouse

In ClickHouse, tables created with the `MergeTree` family of storage engines, including `MergeTree` and `AggregatingMergeTree`, support the `INSERT` operation natively. However, traditional `UPDATE` or `DELETE` operations are not straightforward because ClickHouse is optimised for read and append operations, rather than updates or deletions.

1. **ALTER TABLE ... UPDATE/DELETE**:

   ```sql
   -- Update example
   ALTER TABLE api_logs
   UPDATE third_party_name = 'UpdatedName' WHERE tenantId = 'some-uuid-value';

   -- Delete example
   ALTER TABLE api_logs
   DELETE WHERE tenantId = 'some-uuid-value';
   ```
   It's essential to be aware that these operations aren't instantaneous. They create new parts in the background and merge them, effectively rewriting parts of the table.

2. **ReplacingMergeTree**:

   Another approach before `ALTER TABLE ... UPDATE/DELETE` was introduced using the `ReplacingMergeTree` engine. With `ReplacingMergeTree`, during the merge process, if there are multiple rows with the same primary key, only the latest version (based on a version number or timestamp column) is kept.
For updates, you would re-insert rows with the same primary key but updated values and a newer version number or timestamp. During the next merge, old versions of the record would be replaced by the new ones.

However, even with these mechanisms in place, it's worth noting that frequent updates and deletes can be resource-intensive due to the need to frequently merge parts in ClickHouse. ClickHouse's design favors append-heavy, update-light workloads. If frequent updates or deletions are a primary requirement, other databases might be more suitable.

API timestamp - range queries between a specific time:

SELECT *
FROM api_logs
WHERE api_timestamp BETWEEN '2023-01-01 00:00:00' AND '2023-01-31 23:59:59';

API call completion time:
Count, Filtering, Mean, Median:

SELECT
    COUNT(*) AS total_calls,
    AVG(completion_time_ms) AS mean_completion_time,
    quantile(0.5)(completion_time_ms) AS median_completion_time
FROM api_logs
WHERE completion_time_ms > 500; -- Filtering for calls taking more than 500ms

HTTP status code - Count, Filtering, Error Rate Calculation:

SELECT
    http_status_code,
    COUNT(*) AS total_calls,
    (COUNT(*) * 100) / (SELECT COUNT(*) FROM api_logs) AS error_rate_percentage
FROM api_logs
WHERE http_status_code >= 400
GROUP BY http_status_code;

Count, Filtering, grouping:

SELECT
    url,
    COUNT(*) AS total_calls
FROM api_logs
WHERE url LIKE '%example.com%'
GROUP BY url;

SELECT
    third_party_name,
    COUNT(*) AS total_calls
FROM api_logs
WHERE third_party_name = 'ThirdPartyExample'
GROUP BY third_party_name;

SELECT
    tenantId,
    COUNT(*) AS total_calls
FROM api_logs
WHERE tenantId = 'some-uuid-value'
GROUP BY tenantId;

SELECT
    operation_category,
    operation_subcategory,
    COUNT(*) AS total_calls
FROM api_logs
WHERE operation_category = 'ExampleCategory'
GROUP BY operation_category, operation_subcategory;

SELECT
    entity_type,
    entity_id,
    COUNT(*) AS total_calls
FROM api_logs
WHERE entity_type = 'ExampleType'
GROUP BY entity_type, entity_id;

