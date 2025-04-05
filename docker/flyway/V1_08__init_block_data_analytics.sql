CREATE TABLE IF NOT EXISTS "public"."block_data_analytics" (
        "id" BIGINT GENERATED ALWAYS AS IDENTITY,
        "chain_info_id" INT NOT NULL,
        "height" BIGINT NOT NULL,
        "timestamp" timestamptz NOT NULL,

        "total_txs_bytes" INT NOT NULL,
        "total_gas_used" INT NOT NULL,
        "total_gas_wanted" INT NOT NULL, 
        "success_txs_count" INT NOT NULL,
        "failed_txs_count" INT NOT NULL,

        PRIMARY KEY ("id", "chain_info_id"),
        CONSTRAINT fk_chain_info_id FOREIGN KEY (chain_info_id) REFERENCES meta.chain_info (id) ON DELETE CASCADE ON UPDATE CASCADE
    )
PARTITION BY
    LIST ("chain_info_id");

CREATE INDEX IF NOT EXISTS block_data_analytics_idx_01 ON public.block_data_analytics (height);


CREATE TABLE IF NOT EXISTS "public"."block_message_analytics" (
        "id" BIGINT GENERATED ALWAYS AS IDENTITY,
        "chain_info_id" INT NOT NULL,
        "height" BIGINT NOT NULL,
        "timestamp" timestamptz NOT NULL,

        "message_type_id" INT NOT NULL,
        "success" BOOLEAN NOT NULL,

        PRIMARY KEY ("id", "chain_info_id"),
        CONSTRAINT fk_chain_info_id FOREIGN KEY (chain_info_id) REFERENCES meta.chain_info (id) ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT fk_message_type_id FOREIGN KEY (message_type_id, chain_info_id) REFERENCES meta.message_type (id, chain_info_id)
    )
PARTITION BY
    LIST ("chain_info_id");

CREATE INDEX IF NOT EXISTS block_message_analytics_idx_01 ON public.block_message_analytics (height);
CREATE INDEX IF NOT EXISTS block_message_analytics_idx_02 ON public.block_message_analytics (message_type_id, height);
CREATE INDEX IF NOT EXISTS block_message_analytics_idx_03 ON public.block_message_analytics USING btree (chain_info_id, success, height asc);