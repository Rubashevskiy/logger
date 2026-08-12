CREATE TABLE IF NOT EXISTS {{TABLE_NAME}} (
	data_hash uuid NOT NULL,
	occurence_count int8 DEFAULT 1 NOT NULL,
	"module" text NOT NULL,
	"level" text NOT NULL,
	code int8 NOT NULL,
	msg text NOT NULL,
	debug text NULL,
	upd_dttm timestamptz DEFAULT now() NOT NULL,
	read_flg bool DEFAULT false NOT NULL,
	event_id uuid DEFAULT gen_random_uuid() NOT NULL,
	CONSTRAINT trade_sys_pkey PRIMARY KEY (data_hash)
);

CREATE INDEX IF NOT EXISTS idx_level_time_{{TABLE_SUFFIX}} ON {{TABLE_NAME}} USING btree (level, upd_dttm DESC);
CREATE INDEX IF NOT EXISTS idx_upd_dttm_{{TABLE_SUFFIX}} ON {{TABLE_NAME}} USING btree (upd_dttm DESC);