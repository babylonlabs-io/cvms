package model

import (
	"time"

	"github.com/uptrace/bun"
)

type BlockDataAnalytics struct {
	bun.BaseModel `bun:"table:block_data_analytics"`

	ID              int64     `bun:"id,pk,autoincrement"`
	ChainInfoID     int64     `bun:"chain_info_id,pk,notnull"`
	Height          int64     `bun:"height,notnull"`
	Timestamp       time.Time `bun:"timestamp,notnull"`
	TotalTxsBytes   int64     `bun:"total_txs_bytes"`
	TotalGasUsed    int64     `bun:"total_gas_used"`
	TotalGasWanted  int64     `bun:"total_gas_wanted"`
	SuccessTxsCount int64     `bun:"success_txs_count"`
	FailedTxsCount  int64     `bun:"failed_txs_count"`
}

type BlockMessageAnalytics struct {
	bun.BaseModel `bun:"table:block_message_analytics"`

	ID            int64     `bun:"id,pk,autoincrement"`
	ChainInfoID   int64     `bun:"chain_info_id,pk,notnull"`
	Height        int64     `bun:"height,notnull"`
	Timestamp     time.Time `bun:"timestamp,notnull"`
	MessageTypeID int64     `bun:"message_type_id"`
	Success       bool      `bun:"success"`
}
