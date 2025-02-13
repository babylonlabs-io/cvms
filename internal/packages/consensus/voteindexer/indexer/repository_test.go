package indexer

import (
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/indexer/model"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/cosmostation/cvms/internal/packages/consensus/voteindexer/repository"
	"github.com/stretchr/testify/assert"
)

var (
	p = common.Packager{
		ChainName:    "neutron",
		ChainID:      "pion-1",
		ProtocolType: "cosmos",
		Endpoints: common.Endpoints{
			RPCs: []string{"https://rpc-office.cosmostation.io/neutron-testnet"},
			APIs: []string{"https://lcd-office.cosmostation.io/neutron-testnet"},
		},
		Logger: logger.GetTestLogger(),
	}
)

func TestSetDB(t *testing.T) {
	tempDBName := "temp"
	indexerDB, err := common.NewTestLoaclIndexerDB(tempDBName)
	assert.NoError(t, err)

	p.SetIndexerDB(indexerDB)
	p.SetRetentionTime("1h")
	p.IsConsumerChain = true

	idx, err := NewVoteIndexer(p)
	assert.NoError(t, err)

	err = idx.InitChainInfoID()
	assert.NoError(t, err)

	err = idx.repo.InitPartitionTablesByChainInfoID(repository.IndexName, idx.ChainID, 24256198)
	assert.NoError(t, err)

	err = idx.repo.CreateValidatorInfoPartitionTableByChainID(idx.ChainID)
	assert.NoError(t, err)

	idx.ChainInfoID, err = idx.repo.SelectChainInfoIDByChainID(idx.ChainID)
	assert.NoError(t, err)

	models := []model.ValidatorInfo{
		{
			ChainInfoID:     idx.ChainInfoID,
			HexAddress:      "0x1234567890",
			OperatorAddress: "cosmos1valoper1234567890",
			Moniker:         "Cosmostation",
		},
	}

	err = idx.repo.InsertValidatorInfoList(models)
	assert.NoError(t, err)

	// by assigning consumer key
	new := []model.ValidatorInfo{
		{
			ChainInfoID:     idx.ChainInfoID,
			HexAddress:      "0x234567890",
			OperatorAddress: "cosmos1valoper1234567890",
			Moniker:         "Cosmostation",
		},
	}

	err = idx.repo.InsertValidatorInfoList(new)
	assert.NoError(t, err)
}

/*
	INSERT INTO "meta"."validator_info" ("chain_info_id", "hex_address", "operator_address", "moniker")
	VALUES (1, '0x234567890', 'cosmos1valoper1234567890', 'Cosmostation')
	ON CONFLICT (chain_info_id, operator_address)
	DO UPDATE SET hex_address = EXCLUDED.hex_address;

	_, err = idx.DB.NewInsert().
		Model(&new).
		ExcludeColumn("id").
	On("CONFLICT (chain_info_id, operator_address) DO UPDATE").
	Set("other_field = EXCLUDED.other_field").

		Exec(ctx)
	assert.NoError(t, err)
*/
