package model

import (
	"fmt"

	"github.com/uptrace/bun"
)

type BabylonBTCRoll struct {
	bun.BaseModel    `bun:"table:babylon_btc_lightclient"`
	ID               int64  `bun:"id,pk,autoincrement"`
	ChainInfoID      int64  `bun:"chain_info_id,pk,notnull"`
	Height           int64  `bun:"height"`
	ReporterID       int64  `bun:"reporter_id"`
	RollForwardCount int64  `bun:"roll_forward_count"`
	RollBackCount    int64  `bun:"roll_back_count"`
	BTCHeight        int64  `bun:"btc_height"`
	IsRollBack       bool   `bun:"is_roll_back"`
	BTCHeaders       string `bun:"btc_headers"`
}

func (bbr BabylonBTCRoll) String() string {
	return fmt.Sprintf("BabylonBTCRoll<%d | forward: %d, back: %d, BTC height %d>",
		bbr.Height,
		bbr.RollForwardCount,
		bbr.RollBackCount,
		bbr.BTCHeight,
	)
}
