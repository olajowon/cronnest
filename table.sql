CREATE TABLE hostgroup (
    id SERIAL PRIMARY KEY,
    name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


CREATE TABLE host (
    id SERIAL PRIMARY KEY,
    address text NOT NULL UNIQUE,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    status text NOT NULL DEFAULT 'enabled'::text
);


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


CREATE TABLE operation_record (
    id SERIAL PRIMARY KEY,
    source_type text,
    source_id integer,
    operation_type text,
    data jsonb DEFAULT '{}'::jsonb,
    "user" text,
    created_at timestamp with time zone NOT NULL,
    source_label text
);

