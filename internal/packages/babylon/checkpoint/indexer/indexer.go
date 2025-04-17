package indexer

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"

	"github.com/cosmostation/cvms/internal/common"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/packages/babylon/checkpoint/repository"
)

var subsystem = "babylon_checkpoint"

type CheckpointIndexer struct {
	*common.Indexer
	repo                repository.CheckpointIndexerRepository
	earliestBlockHeight int64
	lastDBEpoch         int64
}

var _ common.IIndexer = (*CheckpointIndexer)(nil)

func NewCheckpointIndexer(p common.Packager) (*CheckpointIndexer, error) {
	status := helper.GetOnChainStatus(p.RPCs, p.ProtocolType)
	if status.ChainID == "" {
		return nil, errors.Errorf("failed to create a new checkpoint indexer: %v", status)
	}
	indexer := common.NewIndexer(p, p.Package, status.ChainID)
	repo := repository.NewRepository(*p.IndexerDB, subsystem, indexertypes.SQLQueryMaxDuration)
	indexer.Lh = indexertypes.LatestHeightCache{LatestHeight: status.BlockHeight}
	return &CheckpointIndexer{indexer, repo, status.EarliestBlockHeight, 0}, nil
}

func (idx *CheckpointIndexer) Start() error {
	err := idx.InitChainInfoID()
	if err != nil {
		return errors.Wrap(err, "failed to init chain_info_id")
	}

	alreadyInit, err := idx.repo.CheckIndexPointerAlreadyInitialized(idx.repo.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to check init tables")
	}
	if !alreadyInit {
		idx.Warnf("it's not initialized in the database, so that this package will initalize at %d as a init index point", idx.Lh.LatestHeight)
		idx.repo.InitPartitionTablesByChainInfoID(idx.repo.IndexName, idx.ChainID, idx.earliestBlockHeight)
	}

	// get last index pointer, index pointer is always initalize if not exist
	initIndexPointer, err := idx.repo.GetLastIndexPointerByIndexTableName(idx.repo.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get last index pointer")
	}

	// get last epoch and save into indexer struct
	lastDBEpoch, err := idx.repo.GetLastEpoch(idx.ChainInfoID)
	if err != nil {
		return errors.WithStack(err)
	}

	err = idx.FetchValidatorInfoList()
	if err != nil {
		return errors.Wrap(err, "failed to fetch validator_info list")
	}

	idx.Infof("loaded last db epoch: %d, index pointer: %d, loaded validator id map: %d", lastDBEpoch, initIndexPointer.Pointer, len(idx.Vim))

	// init indexer metrics
	idx.initLabelsAndMetrics()

	// NOTE: no need to sync lastest height in babylon checkpoint indexer
	// go fetch new height in loop, it must be after init metrics
	// go idx.FetchLatestHeight()
	go idx.Loop(lastDBEpoch)
	// loop partion table time retention by env parameter
	go func() {
		if idx.RetentionPeriod == db.PersistenceMode {
			idx.Infoln("skipped the postgres time retention")
			return
		}
		for {
			idx.Infof("for time retention, delete old records over %s and sleep %s", idx.RetentionPeriod, indexertypes.RetentionQuerySleepDuration)
			idx.repo.DeleteOldValidatorExtensionVoteList(idx.ChainID, idx.RetentionPeriod)
			time.Sleep(indexertypes.RetentionQuerySleepDuration)
		}
	}()
	return nil
}

func (idx *CheckpointIndexer) Loop(indexPoint int64) {
	isUnhealth := false
	for {
		// node health check
		if isUnhealth {
			healthAPIs := healthcheck.FilterHealthEndpoints(idx.APIs, idx.ProtocolType)
			for _, api := range healthAPIs {
				idx.SetAPIEndPoint(api)
				idx.Debugf("API endpoint will be changed with health endpoint for this package: %s", api)
				isUnhealth = false
				break
			}

			healthRPCs := healthcheck.FilterHealthRPCEndpoints(idx.RPCs, idx.ProtocolType)
			for _, rpc := range healthRPCs {
				idx.SetRPCEndPoint(rpc)
				idx.Debugf("RPC endpoint will be changed with health endpoint for this package: %s", rpc)
				isUnhealth = false
				break
			}

			if len(healthAPIs) == 0 || len(healthRPCs) == 0 {
				isUnhealth = true
				idx.Errorln("failed to get any health endpoints from healthcheck filter, retry sleep 10s")
				time.Sleep(indexertypes.UnHealthSleep)
				continue
			}
		}
		// trying to sync with new index pointer height
		newIndexPointer, err := idx.batchSync(indexPoint)
		if err != nil {
			common.Health.With(idx.RootLabels).Set(0)
			common.Ops.With(idx.RootLabels).Inc()
			isUnhealth = true
			idx.Errorf("failed to sync status in %d epoch: %s, it will be retried after sleep %s...",
				indexPoint, err, indexertypes.AfterFailedRetryTimeout.String(),
			)
			time.Sleep(indexertypes.AfterFailedRetryTimeout)
			continue
		}

		// update index point
		indexPoint = newIndexPointer

		// update health and ops
		common.Health.With(idx.RootLabels).Set(1)
		common.Ops.With(idx.RootLabels).Inc()
	}
}

// insert chain-info into chain_info table
func (idx *CheckpointIndexer) InitChainInfoID() error {
	isNewChain := false
	var chainInfoID int64
	chainInfoID, err := idx.repo.SelectChainInfoIDByChainID(idx.ChainID)
	if err != nil {
		if err == sql.ErrNoRows {
			idx.Infof("this is new chain id: %s", idx.ChainID)
			isNewChain = true
		} else {
			return errors.Wrap(err, "failed to select chain_info_id by chain-id")
		}
	}

	if isNewChain {
		chainInfoID, err = idx.repo.InsertChainInfo(idx.ChainName, idx.ChainID, idx.Mainnet)
		if err != nil {
			return errors.Wrap(err, "failed to insert new chain_info_id by chain-id")
		}
	}

	idx.ChainInfoID = chainInfoID
	return nil
}

func (idx *CheckpointIndexer) FetchValidatorInfoList() error {
	// get already saved validator-set list for mapping validators ids
	validatorInfoList, err := idx.repo.GetValidatorInfoListByChainInfoID(idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get validator info list")
	}

	// when the this pacakge starts, set validator-id map
	for _, validator := range validatorInfoList {
		idx.Vim[validator.HexAddress] = int64(validator.ID)
	}

	return nil
}
