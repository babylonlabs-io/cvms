package indexer

import (
	"encoding/json"
	"fmt"
	"time"
)

// SUPPORTED_MESSAGE_TYPES in CVMS
const (
	BabylonInjectedCheckpointMessageType = "/babylon.checkpointing.v1.MsgInjectedCheckpoint"
)

var (
	CurrentEpochQueryPath = "/babylon/epoching/v1/current_epoch"
	EpochQueryPath        = func(epochNumber int64) string {
		return fmt.Sprintf("/babylon/epoching/v1/epochs/%d", epochNumber)
	}

	// TODO: move into common api
	BlockTxsQueryPath = func(blockHeight int64) string {
		return fmt.Sprintf("/cosmos/tx/v1beta1/txs/block/%d?pagination.limit=1", blockHeight)
	}
)

type CurrentEpochResponse struct {
	CurrentEpoch  string `json:"current_epoch"`
	EpochBoundary string `json:"epoch_boundary"`
}

// "epoch":{"epoch_number":"1","current_epoch_interval":"360","first_block_height":"1","last_block_time":"2025-01-08T09:28:50.790836501Z","sealer_app_hash_hex":"24a97c396101ef034ffe1fd0b7d78e22b6a3beec92222a6219162c553198a3c1","sealer_block_hash":"a794387592a2e2eb85d1fc21acc079233dfbb333b6a2ad68337120f3ecfb5aa7"}}
type EpochResponse struct {
	Epoch struct {
		EpochNumber          string `json:"epoch_number"`
		CurrentEpochInterval string `json:"current_epoch_interval"`
		FirstBlockHeight     string `json:"first_block_height"`
		LastBlockTime        string `json:"last_block_time"`
		SealerAppHashHex     string `json:"sealer_app_hash_hex"`
		SealerBlockHash      string `json:"sealer_block_hash"`
	} `json:"epoch"`
}

// MsgInjectedCheckpoint is the structure for the specific message type.
type MsgInjectedCheckpoint struct {
	Ckpt struct {
		Ckpt struct {
			EpochNum    string `json:"epoch_num"`
			BlockHash   string `json:"block_hash"`
			Bitmap      string `json:"bitmap"`
			BlsMultiSig string `json:"bls_multi_sig"`
		} `json:"ckpt"`
		Status    string        `json:"status"`
		BlsAggrPk string        `json:"bls_aggr_pk"`
		PowerSum  string        `json:"power_sum"`
		Lifecycle []interface{} `json:"lifecycle"`
	} `json:"ckpt"`
	ExtendedCommitInfo ExtendedCommitInfo `json:"extended_commit_info"`
}

type ExtendedCommitInfo struct {
	Round int64 `json:"round"`
	Votes []BabylonExtendVote
}

// NOTE: in the future, we consider how to aggregate the extension vote types by being sent ABCI++
//
//	Sample of the extension vote
//	{
//		"validator": {
//		"address": "h4ZQ7nYFnVYQ8RhOmrTLTj/ScHI=",
//		"power": "10000000"
//		},
//		"vote_extension": "CjFiYm52YWxvcGVyMTA5eDRydXNweGFyd3Q2MnB1d2NlbmhjbHczNmw5djdqOTJmMGV4EjFiYm52YWxvcGVyMXM3cjlwbW5rcWt3NHZ5ODNycDhmNGR4dGZjbGF5dXJqamo0ZXVqGiCk+9wsmb8+zbqc4Xci9yK48hy3B873VvBWLcWKsFxtdiACKNAFMjC0rVtg9VKlWdACyQzwxO/aE3VhXq/gnU8Kl6mzWm13skF0/NbXg9+6NkyR/0AZTWc=",
//		"extension_signature": "iW59o/5+KpVOrGz/plUpyBO9jq2XX4cg33hXPaNo26SiX/CSjO6lNPO7rAvpP0l6PQ3u+9bwxSs0LoCBGOX2AA==",
//		"block_id_flag": "BLOCK_ID_FLAG_COMMIT"
//	},
type BabylonExtendVote struct {
	Validator struct {
		Address string `json:"address"`
		Power   string `json:"power"`
	} `json:"validator"`
	VoteExtension      string `json:"vote_extension"`
	ExtensionSignature string `json:"extension_signature"`
	BlockIDFlag        string `json:"block_id_flag"`
}

// TODO: I think this types should move into common cosmos types
type CosmosTx struct {
	Body struct {
		Messages []json.RawMessage `json:"messages"`
	} `json:"body"`
	AuthInfo   interface{} `json:"-"`
	Signatures []string    `json:"-"`
}

type BlockTxsResponse struct {
	Txs   []CosmosTx `json:"txs"`
	Block struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
	} `json:"block"`
}
