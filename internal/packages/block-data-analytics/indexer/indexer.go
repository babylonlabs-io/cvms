package indexer

import (
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"

	"github.com/cosmostation/cvms/internal/common"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/packages/block-data-analytics/repository"
)

var (
	subTableName                     = "block_message_analytics"
	subMetaTableName                 = "message_type"
	_                common.IIndexer = (*BDAIndexer)(nil)
)

// BDAIndexer means BlockDataAnalyticsIndexer
type BDAIndexer struct {
	subsystem           string
	earliestBlockHeight int64
	*common.Indexer
	repository.BDAIndexerRepository
}

func NewBDAIndexer(p common.Packager) (*BDAIndexer, error) {
	subsystem := strings.ReplaceAll(p.Package, "-", "_")
	status := helper.GetOnChainStatus(p.RPCs, p.ProtocolType)
	if status.ChainID == "" {
		return nil, errors.Errorf("failed to create new %s", subsystem)
	}
	indexer := common.NewIndexer(p, subsystem, status.ChainID)
	repo := repository.NewRepository(*p.IndexerDB, subsystem, indexertypes.SQLQueryMaxDuration)
	indexer.Lh = indexertypes.LatestHeightCache{LatestHeight: status.BlockHeight}
	return &BDAIndexer{subsystem, status.EarliestBlockHeight, indexer, repo}, nil
}

func (idx *BDAIndexer) Start() error {
	err := idx.InitChainInfoID()
	if err != nil {
		return errors.Wrap(err, "failed to init chain_info_id")
	}

	err = idx.CreatePartitionTableInMeta(subMetaTableName, idx.ChainID)
	if err != nil {
		return errors.Wrap(err, "failed to init meta partition table")
	}
	err = idx.InitPartitionTablesByChainInfoID(idx.IndexName, idx.ChainID, idx.earliestBlockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to init indexer partition table")
	}
	err = idx.InitPartitionTablesByChainInfoID(subTableName, idx.ChainID, idx.earliestBlockHeight)
	if err != nil {
		return errors.Wrap(err, "failed to init indexer partition table")
	}

	initIndexPointer, err := idx.GetLastIndexPointerByIndexTableName(idx.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get last index pointer")
	}

	err = idx.FetchValidatorInfoList()
	if err != nil {
		return errors.Wrap(err, "failed to fetch validator_info list")
	}

	idx.Infof("loaded index pointer: %d, loaded VIM length: %d VAM: %d", initIndexPointer.Pointer, len(idx.Vim), len(idx.VAM))

	// init indexer metrics
	idx.initMetrics()
	go idx.FetchLatestHeight()
	go idx.Loop(initIndexPointer.Pointer)
	go func() {
		for {
			idx.Infof("for time retention, delete old records over %s and sleep %s", idx.RetentionPeriod, indexertypes.RetentionQuerySleepDuration)
			idx.DeleteOldRecords(idx.ChainID, idx.RetentionPeriod)
			time.Sleep(indexertypes.RetentionQuerySleepDuration)
		}
	}()
	return nil
}

func (idx *BDAIndexer) Loop(indexPoint int64) {
	isUnhealth := false
	for {
		// node health check
		if isUnhealth {
			healthAPIs := healthcheck.FilterHealthEndpoints(idx.APIs, idx.ProtocolType)
			for _, api := range healthAPIs {
				idx.SetAPIEndPoint(api)
				idx.Warnf("API endpoint will be changed with health endpoint for this package: %s", api)
				isUnhealth = false
				break
			}

			healthRPCs := healthcheck.FilterHealthRPCEndpoints(idx.RPCs, idx.ProtocolType)
			for _, rpc := range healthRPCs {
				idx.SetRPCEndPoint(rpc)
				idx.Warnf("RPC endpoint will be changed with health endpoint for this package: %s", rpc)
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
			idx.Errorf("batch sync: %s. it will be retried after sleep %s...", err, indexertypes.AfterFailedRetryTimeout.String())
			time.Sleep(indexertypes.AfterFailedRetryTimeout)
			continue
		}

		// update index point
		indexPoint = newIndexPointer

		// update health and ops
		common.Health.With(idx.RootLabels).Set(1)
		common.Ops.With(idx.RootLabels).Inc()

		// logging & sleep
		if (idx.Lh.LatestHeight) > indexPoint {
			// when node catching_up is true, sleep 100 milli sec
			idx.Infof("updated index pointer is %d ... remaining %d blocks", indexPoint, (idx.Lh.LatestHeight - indexPoint))
			time.Sleep(indexertypes.CatchingUpSleepDuration)
		} else {
			// when node already catched up, sleep 5 sec
			idx.Infof("updated index pointer to %d and sleep %s sec...", indexPoint, indexertypes.DefaultSleepDuration.String())
			time.Sleep(indexertypes.DefaultSleepDuration)
		}
	}
}

func (idx *BDAIndexer) InitChainInfoID() error {
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
	idx.Debugf("set chain info id: %d", chainInfoID)
	return nil
}

func (idx *BDAIndexer) FetchValidatorInfoList() error {
	messageTypeList, err := idx.GetMessageTypeListByChainInfoID(idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get validator info list")
	}
	for _, msgType := range messageTypeList {
		idx.Vim[msgType.MessageType] = msgType.ID
		idx.VAM[msgType.ID] = msgType.MessageType
	}
	return nil
}
