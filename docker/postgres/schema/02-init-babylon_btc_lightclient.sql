CREATE TABLE IF NOT EXISTS "public"."babylon_btc_lightclient" (
        "id" BIGINT GENERATED ALWAYS AS IDENTITY,
        "chain_info_id" INT NOT NULL,
        "height" BIGINT NOT NULL,
        "reporter_id" INT NOT NULL,
        "header_count" SMALLINT NOT NULL, 
        "btc_headers" TEXT NOT NUll,
        PRIMARY KEY ("id", "chain_info_id"),
        CONSTRAINT fk_chain_info_id FOREIGN KEY (chain_info_id) REFERENCES meta.chain_info (id) ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT fk_reporter_id FOREIGN KEY (reporter_id, chain_info_id) REFERENCES meta.validator_info (id, chain_info_id),
        CONSTRAINT uniq_babylon_btc_lightclient_by_height UNIQUE ("chain_info_id","height","reporter_id")
    )
PARTITION BY
    LIST ("chain_info_id");

CREATE INDEX IF NOT EXISTS babylon_btc_lightclient_idx_01 ON public.babylon_btc_lightclient (height);
CREATE INDEX IF NOT EXISTS babylon_btc_lightclient_idx_02 ON public.babylon_btc_lightclient (reporter_id, height);
CREATE INDEX IF NOT EXISTS babylon_btc_lightclient_idx_03 ON public.babylon_btc_lightclient USING btree (chain_info_id, reporter_id, height desc);