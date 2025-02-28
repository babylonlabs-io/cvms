package model

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type BabylonFinalityProviderVote struct {
	bun.BaseModel        `bun:"table:babylon_finality_provider"`
	ID                   int64     `bun:"id,pk,autoincrement"`
	ChainInfoID          int64     `bun:"chain_info_id,pk,notnull"`
	Height               int64     `bun:"height,notnull"`
	FinalityProviderPKID int64     `bun:"finality_provider_pk_id,notnull"`
	Status               int64     `bun:"status,notnull"`
	CreatedTime          time.Time `bun:"timestamp,notnull"`
}

func (bfpv BabylonFinalityProviderVote) String() string {
	return fmt.Sprintf("BabylonFinalityProviderVote<%d %d %d %d %d>",
		bfpv.ID,
		bfpv.ChainInfoID,
		bfpv.Height,
		bfpv.FinalityProviderPKID,
		bfpv.Status,
	)
}
