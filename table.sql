CREATE TABLE hostgroup (
    id integer DEFAULT nextval('hostgroups_id_seq'::regclass) PRIMARY KEY,
    name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE UNIQUE INDEX hostgroup_pkey ON hostgroup(id int4_ops);

CREATE TABLE host (
    id SERIAL PRIMARY KEY,
    address text NOT NULL UNIQUE,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    status text NOT NULL DEFAULT 'enabled'::text
);

CREATE UNIQUE INDEX host_pkey ON host(id int4_ops);
CREATE UNIQUE INDEX host_address_key ON host(address text_ops);

CREATE TABLE hostgroup_host (
    hostgroup_id integer,
    host_id integer
);

CREATE TABLE host_crontab (
    id SERIAL PRIMARY KEY,
    host_id integer,
    tab jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    status text,
    msg text,
    last_succeed timestamp with time zone
);

CREATE UNIQUE INDEX host_crontab_pkey ON host_crontab(id int4_ops);

CREATE TABLE operation_records (
    id integer DEFAULT nextval('operation_record_id_seq'::regclass) PRIMARY KEY,
    source_type text,
    source_id integer,
    operation_type text,
    data jsonb DEFAULT '{}'::jsonb,
    user text,
    created_at timestamp with time zone NOT NULL,
    source_label text
);

CREATE UNIQUE INDEX operation_record_pkey ON operation_records(id int4_ops);