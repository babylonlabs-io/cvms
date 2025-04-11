package indexer

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	commonapi "github.com/cosmostation/cvms/internal/common/api"
	"github.com/cosmostation/cvms/internal/common/indexer/model"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/babylon/covenant-committee/repository"
	"github.com/pkg/errors"
)

var subsystem = "babylon_committee"

type CovenantSignatureIndexer struct {
	*common.Indexer
	csRepo               repository.CovenantSignatureRepository
	btcDelRepo           repository.BtcDelegationRepository
	earliestBlockHeight  int64
	covenantCommitteeMap map[string]int64
}

var _ common.IIndexer = (*CovenantSignatureIndexer)(nil)

func NewCovenantSignatureIndexer(p common.Packager) (*CovenantSignatureIndexer, error) {
	status := helper.GetOnChainStatus(p.RPCs, p.ProtocolType)
	if status.ChainID == "" {
		return nil, errors.Errorf("failed to create a new covenant signature indexer: %v", status)
	}

	indexer := common.NewIndexer(p, p.Package, status.ChainID)

	csRepo := repository.NewCovenantSigRepository(*p.IndexerDB, indexertypes.SQLQueryMaxDuration)
	btcDelRepo := repository.NewBtcDelegationRepository(*p.IndexerDB, indexertypes.SQLQueryMaxDuration)
	indexer.Lh = indexertypes.LatestHeightCache{LatestHeight: status.BlockHeight}
	return &CovenantSignatureIndexer{indexer, csRepo, btcDelRepo, status.EarliestBlockHeight, make(map[string]int64, 0)}, nil
}

func (idx *CovenantSignatureIndexer) Start() error {
	err := idx.InitChainInfoID()
	if err != nil {
		return errors.Wrap(err, "failed to init chain_info_id")
	}

	alreadyInit, err := idx.csRepo.CheckIndexPointerAlreadyInitialized(repository.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to check init tables")
	}
	if !alreadyInit {
		idx.Warnf("it's not initialized in the database, so that this package will initalize at %d as a init index point", idx.Lh.LatestHeight)
		idx.Warnf("%s,%s,%d", repository.IndexName, idx.ChainID, idx.Lh.LatestHeight)
		idx.csRepo.InitPartitionTablesByChainInfoID(repository.IndexName, idx.ChainID, idx.earliestBlockHeight)
		idx.btcDelRepo.CreatePartitionTable(repository.SubIndexName, idx.ChainID)
		idx.csRepo.CreateCovenantCommitteeInfoPartitionTableByChainID(idx.ChainID)
	}

	// get last index pointer, index pointer is always initalize if not exist
	initIndexPointer, err := idx.csRepo.GetLastIndexPointerByIndexTableName(repository.IndexName, idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get last index pointer")
	}

	// initialize babylon covenant committee
	newCovenantCommitteeInfoList := []model.CovenantCommitteeInfo{}
	covenantCommittee, err := commonapi.GetBalbylonCovenantCommiteeParams(idx.CommonClient)
	if err != nil {
		return errors.Wrap(err, "failed to get covenant committee params")
	}
	// request remote covenant mapping name
	remoteCCMonikerMap := idx.getCovenantComitteeMoniker(idx.Mainnet)

	for _, pk := range covenantCommittee {
		moniker := "Unknown"
		if value, ok := remoteCCMonikerMap[pk]; ok {
			moniker = value
		}

		newCovenantCommitteeInfoList = append(newCovenantCommitteeInfoList, model.CovenantCommitteeInfo{
			ChainInfoID:   idx.ChainInfoID,
			CovenantBtcPk: pk,
			Moniker:       moniker,
		})
	}

	err = idx.csRepo.UpsertCovenantCommitteeInfoList(newCovenantCommitteeInfoList)
	if err != nil {
		return errors.Wrap(err, "failed to upsert covenant committee list")
	}
	err = idx.FetchValidatorInfoList()
	if err != nil {
		return errors.Wrap(err, "failed to fetch covenant committee list")
	}

	idx.Infof("loaded last index pointer: %d", initIndexPointer.Pointer)

	// init indexer metrics
	idx.initLabelsAndMetrics()
	idx.initMetricState(idx.covenantCommitteeMap)

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
			idx.csRepo.DeleteOldCovenantSignatureList(idx.ChainID, idx.RetentionPeriod)
			idx.btcDelRepo.DeleteOldBtcDelegationList(idx.ChainID, idx.RetentionPeriod)
			time.Sleep(indexertypes.RetentionQuerySleepDuration)
		}
	}()

	return nil
}

// insert chain-info into chain_info table
func (idx *CovenantSignatureIndexer) InitChainInfoID() error {
	isNewChain := false
	var chainInfoID int64
	chainInfoID, err := idx.csRepo.SelectChainInfoIDByChainID(idx.ChainID)
	if err != nil {
		if err == sql.ErrNoRows {
			idx.Infof("this is new chain id: %s", idx.ChainID)
			isNewChain = true
		} else {
			return errors.Wrap(err, "failed to select chain_info_id by chain-id")
		}
	}

	if isNewChain {
		chainInfoID, err = idx.csRepo.InsertChainInfo(idx.ChainName, idx.ChainID, idx.Mainnet)
		if err != nil {
			return errors.Wrap(err, "failed to insert new chain_info_id by chain-id")
		}
	}

	idx.ChainInfoID = chainInfoID
	return nil
}

func (idx *CovenantSignatureIndexer) Loop(indexPoint int64) {
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
		newIndexPointerHeight := indexPoint + 1

		// trying to sync with new index pointer height
		newIndexPointer, err := idx.batchSync(indexPoint, newIndexPointerHeight)
		if err != nil {
			common.Health.With(idx.RootLabels).Set(0)
			common.Ops.With(idx.RootLabels).Inc()
			isUnhealth = true
			idx.Errorf("failed to sync status in %d block %s, it will be retried after sleep %s...",
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

		// logging & sleep
		if idx.Lh.LatestHeight > indexPoint {
			// when node catching_up is true, sleep 100 milli sec
			idx.WithField("catching_up", true).
				Infof("latest height is %d but updated index pointer is %d ... remaining %d blocks", idx.Lh.LatestHeight, indexPoint, (idx.Lh.LatestHeight - indexPoint))
			time.Sleep(indexertypes.CatchingUpSleepDuration)
		} else {
			// when node already catched up, sleep 5 sec
			idx.WithField("catching_up", false).
				Infof("updated index pointer to %d and sleep %s sec...", indexPoint, indexertypes.DefaultSleepDuration.String())
			time.Sleep(indexertypes.DefaultSleepDuration)
		}
	}
}

func (idx *CovenantSignatureIndexer) FetchValidatorInfoList() error {
	// get already saved covenant committee members list for mapping covenant committee ids
	covenantCommitteeInfoList, err := idx.csRepo.GetCovenantCommitteeInfoListByChainInfoID(idx.ChainInfoID)
	if err != nil {
		return errors.Wrap(err, "failed to get validator info list")
	}

	// when the this pacakge starts, set committee-id map
	for _, committee := range covenantCommitteeInfoList {
		idx.covenantCommitteeMap[committee.CovenantBtcPk] = int64(committee.ID)
	}

	return nil
}

var mainnetDefaultMonikers = map[string]string{
	"d45c70d28f169e1f0c7f4a78e2bc73497afe585b70aa897955989068f3350aaa": "Babylon Labs - Signer 0",
	"4b15848e495a3a62283daaadb3f458a00859fe48e321f0121ebabbdd6698f9fa": "Babylon Labs - Signer 1",
	"23b29f89b45f4af41588dcaf0ca572ada32872a88224f311373917f1b37d08d1": "Babylon Labs - Signer 2",
	"d3c79b99ac4d265c2f97ac11e3232c07a598b020cf56c6f055472c893c0967ae": "CoinSummer Labs",
	"f178fcce82f95c524b53b077e6180bd2d779a9057fdff4255a0af95af918cee0": "RockX",
	"8242640732773249312c47ca7bdb50ca79f15f2ecc32b9c83ceebba44fb74df7": "AltLayer",
	"cbdd028cfe32c1c1f2d84bfec71e19f92df509bba7b8ad31ca6c1a134fe09204": "Zellic",
	"e36200aaa8dce9453567bba108bdc51f7f1174b97a65e4dc4402fc5de779d41c": "Informal Systems",
	"de13fc96ea6899acbdc5db3afaa683f62fe35b60ff6eb723dad28a11d2b12f8c": "Cubist",
}

func (idx *CovenantSignatureIndexer) getCovenantComitteeMoniker(mainnet bool) map[string]string {
	url := BabylonCovenantCommitteeMonikerFromTestnet
	if mainnet {
		url = BabylonCovenantCommitteeMonikerFromMainnet
	}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		idx.Warnln("failed to fetch remote covenant committee moniker, using default monikers")
		if mainnet {
			return mainnetDefaultMonikers
		}
		return map[string]string{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		idx.Errorf("failed to read response body for covenant committee moniker: %s", err)
		return map[string]string{}
	}

	var remoteCCMonikerMap map[string]string
	if err := json.Unmarshal(body, &remoteCCMonikerMap); err != nil {
		idx.Errorf("failed to unmarshal covenant committee moniker JSON: %s", err)
		return map[string]string{}
	}

	return remoteCCMonikerMap
}
