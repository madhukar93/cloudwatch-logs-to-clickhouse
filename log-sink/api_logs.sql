CREATE TABLE api_logs
(
    api_timestamp DateTime,
    -- event_id is a unique identifier for each API call, used as primary key
    event_id UUID,
    completion_time_ms UInt32,
    url String,
    http_status_code UInt16,
    third_party_name String,
    tenantId UUID,
    entity_type String,
    entity_id String,
    operation_category String,
    operation_subcategory String
    -- minmax index for http_status_code will optimize for range queries
    INDEX http_status_code_index http_status_code TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY tenantId
// probably remove tenantID here, since we’re partitioning by tenant ID.
ORDER BY (api_timestamp, event_id);
