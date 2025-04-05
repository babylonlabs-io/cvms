CREATE TABLE IF NOT EXISTS "public"."babylon_finality_provider" (
        "id" BIGINT GENERATED ALWAYS AS IDENTITY,
        "chain_info_id" INT NOT NULL,
        "height" BIGINT NOT NULL,
        "finality_provider_pk_id" INT NOT NULL,
        "status" SMALLINT NOT NULL, 
        "timestamp" timestamptz NOT NULL,
        PRIMARY KEY ("id", "chain_info_id"),
        CONSTRAINT fk_chain_info_id FOREIGN KEY (chain_info_id) REFERENCES meta.chain_info (id) ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT fk_finality_provider_pk_id FOREIGN KEY (finality_provider_pk_id, chain_info_id) REFERENCES meta.finality_provider_info (id, chain_info_id),
        CONSTRAINT uniq_babylon_finality_provider_by_height UNIQUE ("chain_info_id","height","finality_provider_pk_id")
    )
PARTITION BY
    LIST ("chain_info_id");

CREATE INDEX IF NOT EXISTS babylon_finality_provider_idx_01 ON public.babylon_finality_provider (height);
CREATE INDEX IF NOT EXISTS babylon_finality_provider_idx_02 ON public.babylon_finality_provider (finality_provider_pk_id, height);
CREATE INDEX IF NOT EXISTS babylon_finality_provider_idx_03 ON public.babylon_finality_provider USING btree (chain_info_id, finality_provider_pk_id, height asc);