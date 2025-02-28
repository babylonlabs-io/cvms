package indexer

import (
	"database/sql"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/babylon/btc-lightclient/repository"
	"github.com/pkg/errors"
)

var (
	subsystem                 = "babylon_btc_lightclient"
	_         common.IIndexer = (*BTCLightClientIndexer)(nil)
)

type BTCLightClientIndexer struct {
	*common.Indexer
	repository.BTCLightClientIndexerRepository
	earliestBlockHeight int64
}

func NewBTCLightClientIndexer(p common.Packager) (*BTCLightClientIndexer, error) {
	status := helper.GetOnChainStatus(p.RPCs, p.ProtocolType)
	if status.ChainID == "" {
		return nil, errors.Errorf("failed to create a new indexer: %v", status)
	}
	indexer := common.NewIndexer(p, p.Package, status.ChainID)
	repo := repository.NewRepository(*p.IndexerDB, subsystem, indexertypes.SQLQueryMaxDuration)
	indexer.Lh = indexertypes.LatestHeightCache{LatestHeight: status.BlockHeight}
	idx := &BTCLightClientIndexer{indexer, repo, status.EarliestBlockHeight}
	idx.initMetrics()
	return idx, nil
}

// NOTE: Babylon operates a Bitcoin light client that stores Bitcoin chain headers.
// The light client is maintained up to date by the Vigilante (more specifically, the Vigilante Reporter), which tracks BTC state and submits the latest BTC headers to Babylon.
// Ensuring that the BTC light client is up to date and can recover from Bitcoin re-orgs is very important for both the BTC Staking and the BTC Timestamping protocols.
// The CVMS tool will index the Babylon node events BTCRollForward and BTCRollBack
func (idx *BTCLightClientIndexer) Start() error {
	err := idx.InitChainInfoID()
	if err != nil {
		return errors.Wrap(err, "failed to init chain_info_id")
	}

	alreadyInit, err := idx.CheckIndexPointerAlreadyInitialized(idx.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to check init tables")
	}
	if !alreadyInit {
		idx.Infof("the package will initalize at %d as a init index point", idx.Lh.LatestHeight)
		idx.InitPartitionTablesByChainInfoID(idx.IndexName, idx.ChainID, idx.earliestBlockHeight)
		idx.CreateVigilanteInfoPartitionTableByChainID(idx.ChainID)
	}

	// get last index pointer, index pointer is always initalize if not exist
	initIndexPointer, err := idx.GetLastIndexPointerByIndexTableName(idx.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get last index pointer")
	}

	err = idx.FetchValidatorInfoList()
	if err != nil {
		return errors.Wrap(err, "failed to fetch validator_info list")
	}

	idx.Infof("loaded index pointer(last saved height): %d", initIndexPointer.Pointer)

	go idx.FetchLatestHeight()
	go idx.Loop(initIndexPointer.Pointer)
	go func() {
		if idx.RetentionPeriod == db.PersistenceMode {
			idx.Infoln("skipped the postgres time retention")
			return
		}
		for {
			idx.Infof("for time retention, delete old records over %s and sleep %s", idx.RetentionPeriod, indexertypes.RetentionQuerySleepDuration)
			idx.DeleteOldRecords(idx.ChainID, idx.RetentionPeriod)
			time.Sleep(indexertypes.RetentionQuerySleepDuration)
		}
	}()
	return nil
}

func (idx *BTCLightClientIndexer) Loop(indexPoint int64) {
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
		if idx.Lh.LatestHeight > indexPoint {
			// when node catching_up is true, sleep 100 milli sec
			idx.Infof("updated index pointer is %d ... remaining %d blocks", indexPoint, (idx.Lh.LatestHeight - indexPoint))
			time.Sleep(indexertypes.CatchingUpSleepDuration)
		} else {
			// when node already catched up, sleep 5 sec
			idx.Infof("updated index pointer to %d and sleep %s sec...", indexPoint, indexertypes.DefaultSleepDuration.String())
		}
	}
}

// insert chain-info into chain_info table
func (idx *BTCLightClientIndexer) InitChainInfoID() error {
	isNewChain := false
	var chainInfoID int64
	chainInfoID, err := idx.SelectChainInfoIDByChainID(idx.ChainID)
	if err != nil {
		if err == sql.ErrNoRows {
			idx.Infof("this is new chain id: %s", idx.ChainID)
			isNewChain = true
		} else {
			return errors.Wrap(err, "failed to select chain_info_id by chain-id")
		}
	}

	if isNewChain {
		chainInfoID, err = idx.InsertChainInfo(idx.ChainName, idx.ChainID, idx.Mainnet)
		if err != nil {
			return errors.Wrap(err, "failed to insert new chain_info_id by chain-id")
		}
	}

	idx.ChainInfoID = chainInfoID
	return nil
}

func (idx *BTCLightClientIndexer) FetchValidatorInfoList() error {
	// get already saved validator-set list for mapping validators ids
	vigilanteInfoList, err := idx.GetVigilanteInfoListByChainInfoID(idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get validator info list")
	}

	// when the this pacakge starts, set validator-id map
	for _, validator := range vigilanteInfoList {
		idx.Vim[validator.OperatorAddress] = int64(validator.ID)
	}

	return nil
}
