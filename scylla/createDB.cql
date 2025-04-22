/* Создаём пространство ключей chrds */
CREATE KEYSPACE chrds WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': '2'}  AND durable_writes = true;

/* Создаём пользователя chrds (Password: CHRDSpass5) для работы приложений */
CREATE ROLE chrds with SUPERUSER = false AND LOGIN = true and PASSWORD = 'CHRDSpass5';

/* Выдаём необходимые гранты */
GRANT ALL ON KEYSPACE chrds TO chrds;
GRANT EXECUTE ON ALL FUNCTIONS TO chrds;
GRANT SELECT ON system.cluster_status TO chrds;

/* Создаём табличную структуру */
CREATE TABLE chrds.dashboard (
    user_id uuid,
    id uuid,
    name text,
    start_time bigint,
    status int,
    stop_time bigint,
    PRIMARY KEY (user_id, id)
) WITH CLUSTERING ORDER BY (id ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.dashboard_chartdata (
    dashboard_id uuid,
    graph_order int,
    id uuid,
    graph_color text,
    graph_func text,
    graph_type text,
    metric text,
    space_id uuid,
    status int,
    PRIMARY KEY (dashboard_id, graph_order, id)
) WITH CLUSTERING ORDER BY (graph_order ASC, id ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.owner_id_user (
    owner_id uuid,
    user_name text,
    PRIMARY KEY (owner_id, user_name)
) WITH CLUSTERING ORDER BY (user_name ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.problem (
    space_id uuid,
    module_id uuid,
    metric text,
    event_time bigint,
    event_time_start bigint,
    status text,
    value text,
    PRIMARY KEY (space_id, module_id, metric)
) WITH CLUSTERING ORDER BY (module_id ASC, metric ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_aggregation (
    space_id uuid,
    object text,
    metric text,
    synt_key text,
    count int,
    max_value float,
    min_value float,
    sum_value float,
    PRIMARY KEY ((space_id, object, metric, synt_key))
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_data01 (
    space_id uuid,
    synt_key text,
    event_time bigint,
    create_time bigint,
    metric text,
    object text,
    status int,
    value float,
    PRIMARY KEY ((space_id, synt_key), event_time)
) WITH CLUSTERING ORDER BY (event_time DESC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_data02 (
    space_id uuid,
    metric text,
    synt_key text,
    event_time bigint,
    create_time bigint,
    object text,
    status int,
    value float,
    PRIMARY KEY ((space_id, metric, synt_key), event_time)
) WITH CLUSTERING ORDER BY (event_time DESC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_data03 (
    space_id uuid,
    synt_key text,
    object text,
    event_time bigint,
    create_time bigint,
    metric text,
    status int,
    value float,
    PRIMARY KEY ((space_id, synt_key, object), event_time)
) WITH CLUSTERING ORDER BY (event_time DESC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_key (
    space_id uuid,
    type_data text,
    flag int,
    synt_key text,
    PRIMARY KEY ((space_id, type_data, flag), synt_key)
) WITH CLUSTERING ORDER BY (synt_key DESC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_text01 (
    space_id uuid,
    synt_key text,
    event_time bigint,
    create_time bigint,
    metric text,
    object text,
    status int,
    value text,
    PRIMARY KEY ((space_id, synt_key), event_time)
) WITH CLUSTERING ORDER BY (event_time DESC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.raw_text02 (
    space_id uuid,
    metric text,
    synt_key text,
    event_time bigint,
    create_time bigint,
    object text,
    status int,
    value text,
    PRIMARY KEY ((space_id, metric, synt_key), event_time)
) WITH CLUSTERING ORDER BY (event_time ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.registration (
    id uuid PRIMARY KEY,
    description text,
    status int,
    type text,
    user_id uuid
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.space (
    id uuid PRIMARY KEY,
    description text,
    status int,
    user_id uuid
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.space_metric_data (
    space_id uuid,
    metric text,
    PRIMARY KEY (space_id, metric)
) WITH CLUSTERING ORDER BY (metric ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.space_metric_text (
    space_id uuid,
    metric text,
    PRIMARY KEY (space_id, metric)
) WITH CLUSTERING ORDER BY (metric ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.task (
    module_id uuid,
    space_id uuid,
    object text,
    metric text,
    critical text,
    data_type text,
    int_id bigint,
    interval bigint,
    status int,
    warning text,
    PRIMARY KEY (module_id, space_id, object, metric)
) WITH CLUSTERING ORDER BY (space_id ASC, object ASC, metric ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.user (
    user_name text PRIMARY KEY,
    description text,
    first_name text,
    last_name text,
    login_time bigint,
    owner_id uuid,
    password text,
    role text,
    status int,
    user_id uuid
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.user_id (
    user_id uuid PRIMARY KEY,
    user_name text
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.user_registration (
    user_id uuid,
    registration_id uuid,
    PRIMARY KEY (user_id, registration_id)
) WITH CLUSTERING ORDER BY (registration_id ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

CREATE TABLE chrds.user_space (
    user_id uuid,
    space_id uuid,
    PRIMARY KEY (user_id, space_id)
) WITH CLUSTERING ORDER BY (space_id ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1.0
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE';

/* Создаём преднастроенного суперпользователя (username: superadmin, password: superadmin) */
INSERT INTO chrds.user ("user_name", "description", "first_name", "last_name", "password", "role", "status", "user_id") VALUES ('superadmin', 'Monitoring System Administrator', 'Admin', 'SuperAdmin', '186cf774c97b60a1c106ef718d10970a6a06e06bef89553d9ae65d938a886eae', 'superadmin', 1, 4467f7dd-9787-4113-8e4a-abdfd83956c6);
INSERT INTO chrds.user_id ("user_id", "user_name") VALUES (4467f7dd-9787-4113-8e4a-abdfd83956c6, 'superadmin');

/* Создаём преднастроенное пространство и модуль для взаимодействия UIManager и DataManager */
INSERT INTO chrds.registration (id, type, status, description, user_id) VALUES (623a57c2-3df5-4287-ada5-82f7e4a0b5db, 'uimanager', 1, 'UIManager', 4467f7dd-9787-4113-8e4a-abdfd83956c6);
INSERT INTO chrds.user_registration (user_id, registration_id) VALUES (4467f7dd-9787-4113-8e4a-abdfd83956c6, 623a57c2-3df5-4287-ada5-82f7e4a0b5db);
INSERT INTO chrds.space (id, description, status, user_id) VALUES (17f0bd20-41cf-4801-a481-ff721a41fa93, 'UIManager', 1, 4467f7dd-9787-4113-8e4a-abdfd83956c6);
INSERT INTO chrds.user_space (user_id, space_id) VALUES (4467f7dd-9787-4113-8e4a-abdfd83956c6, 17f0bd20-41cf-4801-a481-ff721a41fa93);