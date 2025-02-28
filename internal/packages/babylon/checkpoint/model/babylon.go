package model

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type BabylonVoteExtension struct {
	bun.BaseModel         `bun:"table:babylon_checkpoint"`
	ID                    int64     `bun:"id,pk,autoincrement"`
	ChainInfoID           int64     `bun:"chain_info_id,pk,notnull"`
	Epoch                 int64     `bun:"epoch,notnull"`
	Height                int64     `bun:"height,notnull"`
	Timestamp             time.Time `bun:"timestamp,notnull"`
	ValidatorHexAddressID int64     `bun:"validator_hex_address_id,notnull"`
	Status                int64     `bun:"status,notnull"`
}

func (bve BabylonVoteExtension) String() string {
	return fmt.Sprintf("BabylonExtensionVote<%d %d %d %d %d %d %d>",
		bve.ID,
		bve.ChainInfoID,
		bve.Epoch,
		bve.Height,
		bve.Timestamp.Unix(),
		bve.ValidatorHexAddressID,
		bve.Status,
	)
}

// var BlockIDFlag_value = map[string]int32{
// 	"BLOCK_ID_FLAG_UNKNOWN": 0,
// 	"BLOCK_ID_FLAG_ABSENT":  1,
// 	"BLOCK_ID_FLAG_COMMIT":  2,
// 	"BLOCK_ID_FLAG_NIL":     3,
// }

type TotalBabylonVoteExtensionByMoniker struct {
	Moniker      string `bun:"moniker"`
	UnknownCount int32  `bun:"unknown"`
	AbsentCount  int32  `bun:"absent"`
	CommitCount  int32  `bun:"commit"`
	NilCount     int32  `bun:"nil"`
}

func (model TotalBabylonVoteExtensionByMoniker) String() string {
	return fmt.Sprintf("Current Babylon BLS Vote<moniker:%s, unkonw: %d, absent: %d, commit: %d, nil: %d>",
		model.Moniker,
		model.UnknownCount,
		model.AbsentCount,
		model.CommitCount,
		model.NilCount,
	)
}
