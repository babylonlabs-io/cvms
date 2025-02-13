package model

import (
	"fmt"

	"github.com/uptrace/bun"
)

type BabylonBTCRoll struct {
	bun.BaseModel `bun:"table:babylon_btc_lightclient"`
	ID            int64  `bun:"id,pk,autoincrement"`
	ChainInfoID   int64  `bun:"chain_info_id,pk,notnull"`
	Height        int64  `bun:"height,notnull"`
	ReporterID    int64  `bun:"reporter_id,notnull"`
	HeaderCount   int    `bun:"header_count,notnull"`
	BTCHeaders    string `bun:"btc_headers,notnull"`
}

func (bbr BabylonBTCRoll) String() string {
	return fmt.Sprintf("BabylonBTCRoll<%d %d %d %d %s>",
		bbr.ChainInfoID,
		bbr.Height,
		bbr.ReporterID,
		bbr.HeaderCount,
		bbr.BTCHeaders,
	)
}
