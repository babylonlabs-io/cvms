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
	"github.com/cosmostation/cvms/internal/packages/babylon/finality-provider/repository"
)

type FinalityProviderIndexer struct {
	*common.Indexer
	repo repository.FinalityProviderIndexerRepository
}

// Compile-time Assertion
var _ common.IIndexer = (*FinalityProviderIndexer)(nil)

func NewFinalityProviderIndexer(p common.Packager) (*FinalityProviderIndexer, error) {
	status := helper.GetOnChainStatus(p.RPCs, p.ProtocolType)
	if status.ChainID == "" {
		return nil, errors.Errorf("failed to create new voteindexer by failing getting onchain status through %v", p.RPCs)
	}
	indexer := common.NewIndexer(p, p.Package, status.ChainID)
	repo := repository.NewRepository(*p.IndexerDB, indexertypes.SQLQueryMaxDuration)
	indexer.Lh = indexertypes.LatestHeightCache{LatestHeight: status.BlockHeight}
	return &FinalityProviderIndexer{indexer, repo}, nil
}

func (idx *FinalityProviderIndexer) Start() error {
	err := idx.InitChainInfoID()
	if err != nil {
		return errors.Wrap(err, "failed to init chain_info_id")
	}

	alreadyInit, err := idx.repo.CheckIndexPointerAlreadyInitialized(repository.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to check init tables")
	}
	if !alreadyInit {
		idx.Warnf("it's not initialized in the database, so that this package will initalize at %d as a init index point", idx.Lh.LatestHeight)
		idx.repo.InitPartitionTablesByChainInfoID(repository.IndexName, idx.ChainID, idx.Lh.LatestHeight)
		idx.repo.CreateFinalityProviderInfoPartitionTableByChainID(idx.ChainID)
	}

	// get last index pointer, index pointer is always initalize if not exist
	initIndexPointer, err := idx.repo.GetLastIndexPointerByIndexTableName(repository.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get last index pointer")
	}

	err = idx.FetchValidatorInfoList()
	if err != nil {
		return errors.Wrap(err, "failed to fetch validator_info list")
	}

	idx.Infof("loaded index pointer(last saved height): %d", initIndexPointer.Pointer)
	idx.Infof("initial vim length: %d for %s chain", len(idx.Vim), idx.ChainID)

	// go fetch new height in loop, it must be after init metrics
	go idx.FetchLatestHeight()
	// loop
	go idx.Loop(initIndexPointer.Pointer)
	// loop partion table time retention by env parameter
	go func() {
		if idx.RetentionPeriod == db.PersistenceMode {
			idx.Infoln("skipped the postgres time retention")
			return
		}
		for {
			idx.Infof("for time retention, delete old records over %s and sleep %s", idx.RetentionPeriod, indexertypes.RetentionQuerySleepDuration)
			idx.repo.DeleteOldFinalityProviderVoteList(idx.ChainID, idx.RetentionPeriod)
			time.Sleep(indexertypes.RetentionQuerySleepDuration)
		}
	}()
	return nil
}

func (idx *FinalityProviderIndexer) Loop(indexPoint int64) {
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
			idx.Errorf("failed to sync in %d height: %s, it will be retried after sleep %s...", indexPoint, err, indexertypes.AfterFailedFetchSleepDuration.String())
			time.Sleep(indexertypes.AfterFailedRetryTimeout)
			continue
		}

		// update index point
		indexPoint = newIndexPointer

		// update health and ops
		common.Health.With(idx.RootLabels).Set(1)
		common.Ops.With(idx.RootLabels).Inc()

		// logging & sleep
		if (idx.Lh.LatestHeight - finalitySigTimeout) > indexPoint {
			// when node catching_up is true, sleep 100 milli sec
			idx.Infof("updated index pointer is %d ... remaining %d blocks", indexPoint, (idx.Lh.LatestHeight - indexPoint))
			time.Sleep(indexertypes.CatchingUpSleepDuration)
		} else {
			// when node already catched up, sleep 5 sec
			idx.Infof("updated index pointer to %d and sleep %s", indexPoint, indexertypes.DefaultSleepDuration.String())
			time.Sleep(indexertypes.DefaultSleepDuration)
		}
	}
}

// TODO: move into metarepo
// insert chain-info into chain_info table
func (idx *FinalityProviderIndexer) InitChainInfoID() error {
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

// NOTE: in finality provider, validator info means fp info
func (idx *FinalityProviderIndexer) FetchValidatorInfoList() error {
	// get already saved validator-set list for mapping validators ids
	fpInfoList, err := idx.repo.GetFinalityProviderInfoListByChainInfoID(idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get validator info list")
	}

	for _, fp := range fpInfoList {
		idx.Vim[fp.BTCPKs] = int64(fp.ID)
	}

	return nil
}
