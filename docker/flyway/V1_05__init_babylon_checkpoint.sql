CREATE TABLE IF NOT EXISTS "public"."babylon_checkpoint" (
        "id" BIGINT GENERATED ALWAYS AS IDENTITY,
        "chain_info_id" INT NOT NULL,
        "epoch" INT NOT NULL,
        "height" BIGINT NOT NULL,
        "timestamp" timestamptz NOT NULL,
        "validator_hex_address_id" INT NOT NULL,
        "status" SMALLINT NOT NULL, 
        PRIMARY KEY ("id", "chain_info_id"),
        CONSTRAINT fk_chain_info_id FOREIGN KEY (chain_info_id) REFERENCES meta.chain_info (id) ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT fk_validator_hex_address_id FOREIGN KEY (validator_hex_address_id, chain_info_id) REFERENCES meta.validator_info (id, chain_info_id),
        CONSTRAINT uniq_babylon_checkpoint UNIQUE ("chain_info_id","height","validator_hex_address_id")
    )
PARTITION BY
    LIST ("chain_info_id");

CREATE INDEX IF NOT EXISTS babylon_checkpoint_idx_01 ON public.babylon_checkpoint (height);
CREATE INDEX IF NOT EXISTS babylon_checkpoint_idx_02 ON public.babylon_checkpoint (validator_hex_address_id, height);
CREATE INDEX IF NOT EXISTS babylon_checkpoint_idx_03 ON public.babylon_checkpoint USING btree (chain_info_id, validator_hex_address_id, height asc);